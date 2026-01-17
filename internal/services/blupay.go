package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/victtorkaiser/server-apis/internal/config"
	"github.com/victtorkaiser/server-apis/internal/dto"
	"github.com/victtorkaiser/server-apis/internal/models"
	"github.com/victtorkaiser/server-apis/internal/queue"
	"gorm.io/gorm"
)

type BluPayService struct {
	db       *gorm.DB
	redis    *redis.Client
	rabbitMQ *queue.RabbitMQ
	cfg      *config.Config
}

func NewBluPayService(db *gorm.DB, redis *redis.Client, rabbitMQ *queue.RabbitMQ, cfg *config.Config) *BluPayService {
	return &BluPayService{
		db:       db,
		redis:    redis,
		rabbitMQ: rabbitMQ,
		cfg:      cfg,
	}
}

func (s *BluPayService) CreatePayment(ctx context.Context, req *dto.BluPayRequest) (*dto.BluPayResponse, error) {
	// Gera placa aleatória para referência externa se não fornecida
	if req.ExternalRef == "" {
		req.ExternalRef = fmt.Sprintf("ORD-%s", s.generatePlaca())
	}

	// Gera dados automaticamente se não fornecidos
	if err := s.fillMissingData(req); err != nil {
		log.Printf("⚠️ Erro ao gerar dados automáticos: %v (continuando com dados fornecidos)", err)
	}

	// Cria customer
	customer := &models.Customer{
		Name:     req.Name,
		Email:    req.Email,
		Document: req.Document,
		Phone:    req.Phone,
		Country:  "BR",
	}

	if err := s.db.Create(customer).Error; err != nil {
		return nil, fmt.Errorf("erro ao criar customer: %w", err)
	}

	// Cria tracking parameters se existirem
	var trackingParamID *uuid.UUID
	if len(req.UTMParams) > 0 {
		trackingParam := s.mapTrackingParams(req.UTMParams)
		if err := s.db.Create(&trackingParam).Error; err != nil {
			return nil, fmt.Errorf("erro ao criar tracking params: %w", err)
		}
		trackingParamID = &trackingParam.ID
	}

	// Chama API BluPay
	externalResp, err := s.callBluPayAPI(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar API BluPay: %w", err)
	}

	// Extrai dados do PIX
	pixCode := externalResp.Pix.QRCode
	qrCodeURL := s.generateQRCodeURL(pixCode)

	// Cria order
	order := &models.Order{
		TransactionID:       externalResp.ID,
		Status:              models.OrderStatusPending,
		Amount:              req.Amount,
		PaymentMethod:       "pix",
		Platform:            "BluPay",
		PixCode:             pixCode,
		CustomerID:          customer.ID,
		TrackingParameterID: trackingParamID,
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, fmt.Errorf("erro ao criar order: %w", err)
	}

	// Publica evento na fila (se RabbitMQ estiver disponível)
	if s.rabbitMQ != nil {
		s.rabbitMQ.Publish("payment.created", map[string]interface{}{
			"order_id":       order.ID,
			"transaction_id": order.TransactionID,
			"amount":         order.Amount,
			"platform":       "BluPay",
		})
	}

	// Envia para Utmify (pendente)
	go s.sendToUtmifyPending(order, customer, req, trackingParamID, externalResp.Fee.EstimatedFee)

	return &dto.BluPayResponse{
		Success:   true,
		Token:     order.TransactionID,
		PixCode:   pixCode,
		QRCodeURL: qrCodeURL,
		Amount:    order.Amount,
		Nome:      customer.Name,
		CPF:       customer.Document,
		ExpiraEm:  s.formatExpiresAt(externalResp.Pix.ExpiresAt),
	}, nil
}

func (s *BluPayService) callBluPayAPI(req *dto.BluPayRequest) (*dto.BluPayAPIResponse, error) {
	// Prepara metadata com UTM params
	metadata := make(map[string]string)
	if req.UTMParams != nil {
		for k, v := range req.UTMParams {
			if str, ok := v.(string); ok {
				metadata[k] = str
			}
		}
	}
	metadata["orderId"] = req.ExternalRef

	// Prepara payload conforme API BluPay
	payload := dto.BluPayAPIRequest{
		Amount:        req.Amount,
		PaymentMethod: "pix",
		ExternalRef:   req.ExternalRef,
		Customer: dto.BluPayCustomer{
			Name:  req.Name,
			Email: req.Email,
			Phone: req.Phone,
			Document: dto.BluPayDocument{
				Type:   "cpf",
				Number: req.Document,
			},
		},
		Items: []dto.BluPayItem{
			{
				Title:     s.cfg.BluPayProductName,
				UnitPrice: req.Amount,
				Quantity:  1,
				Tangible:  false,
			},
		},
		PostbackUrl:   s.cfg.BluPayWebhookURL,
		WebhookSecret: s.cfg.BluPayWebhookSecret,
		Metadata:      metadata,
	}

	body, _ := json.Marshal(payload)
	log.Printf("📤 [BluPay] Request: %s", string(body))

	httpReq, err := http.NewRequest("POST", s.cfg.BluPayAPIURL+"/transactions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Autenticação Basic Auth (secretKey:publicKey)
	auth := base64.StdEncoding.EncodeToString([]byte(s.cfg.BluPaySecretKey + ":" + s.cfg.BluPayPublicKey))
	httpReq.Header.Set("Authorization", "Basic "+auth)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📡 [BluPay] Response HTTP %d: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("erro na API BluPay: status %d - %s", resp.StatusCode, string(respBody))
	}

	var result dto.BluPayAPIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta BluPay: %w", err)
	}

	if result.ID == "" {
		return nil, fmt.Errorf("ID não encontrado na resposta da API BluPay")
	}

	return &result, nil
}

func (s *BluPayService) generateQRCodeURL(pixCode string) string {
	if pixCode != "" {
		return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s", pixCode)
	}
	return ""
}

func (s *BluPayService) formatExpiresAt(expiresAt string) string {
	// Formato padrão: "2 dias"
	t, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return "2 dias"
	}

	diff := time.Until(t)
	days := int(diff.Hours() / 24)
	if days <= 0 {
		return "24 horas"
	}
	if days == 1 {
		return "1 dia"
	}
	return fmt.Sprintf("%d dias", days)
}

func (s *BluPayService) generatePlaca() string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	placa := ""
	for i := 0; i < 3; i++ {
		placa += string(letters[rand.Intn(len(letters))])
	}
	for i := 0; i < 4; i++ {
		placa += fmt.Sprintf("%d", rand.Intn(10))
	}
	return placa
}

func (s *BluPayService) mapTrackingParams(params map[string]interface{}) models.TrackingParameter {
	tp := models.TrackingParameter{}

	if v, ok := params["utm_source"].(string); ok {
		tp.UtmSource = v
		tp.Src = v
	}
	if v, ok := params["utm_campaign"].(string); ok {
		tp.UtmCampaign = v
	}
	if v, ok := params["utm_medium"].(string); ok {
		tp.UtmMedium = v
	}
	if v, ok := params["utm_content"].(string); ok {
		tp.UtmContent = v
	}
	if v, ok := params["utm_term"].(string); ok {
		tp.UtmTerm = v
	}
	if v, ok := params["sck"].(string); ok {
		tp.Sck = v
	}
	if v, ok := params["xcod"].(string); ok {
		tp.Xcod = v
	}
	if v, ok := params["fbclid"].(string); ok {
		tp.Fbclid = v
	}
	if v, ok := params["gclid"].(string); ok {
		tp.Gclid = v
	}
	if v, ok := params["ttclid"].(string); ok {
		tp.Ttclid = v
	}

	return tp
}

func (s *BluPayService) sendToUtmifyPending(order *models.Order, customer *models.Customer, req *dto.BluPayRequest, trackingParamID *uuid.UUID, gatewayFee int) {
	// Publica na fila para processar de forma assíncrona (se RabbitMQ estiver disponível)
	if s.rabbitMQ != nil {
		s.rabbitMQ.Publish("utmify.pending", map[string]interface{}{
			"order_id":              order.ID,
			"transaction_id":        order.TransactionID,
			"customer_id":           customer.ID,
			"tracking_parameter_id": trackingParamID,
			"platform":              "BluPay",
			"gateway_fee":           gatewayFee,
		})
	}

	// Envia diretamente para Utmify
	if s.cfg.UtmifyAPIURL == "" || s.cfg.UtmifyToken == "" {
		log.Println("⚠️ Utmify não configurado, pulando envio")
		return
	}

	// Prepara tracking parameters (todos os campos são obrigatórios no Utmify)
	trackingParams := map[string]interface{}{
		"utm_source":   getStringOrNull(req.UTMParams, "utm_source"),
		"utm_medium":   getStringOrNull(req.UTMParams, "utm_medium"),
		"utm_campaign": getStringOrNull(req.UTMParams, "utm_campaign"),
		"utm_content":  getStringOrNull(req.UTMParams, "utm_content"),
		"utm_term":     getStringOrNull(req.UTMParams, "utm_term"),
		"src":          getStringOrNull(req.UTMParams, "src"),
		"sck":          getStringOrNull(req.UTMParams, "sck"),
		"xcod":         getStringOrNull(req.UTMParams, "xcod"),
		"fbclid":       getStringOrNull(req.UTMParams, "fbclid"),
		"gclid":        getStringOrNull(req.UTMParams, "gclid"),
		"ttclid":       getStringOrNull(req.UTMParams, "ttclid"),
	}

	// Monta payload para Utmify
	utmifyPayload := map[string]interface{}{
		"orderId":       order.TransactionID,
		"platform":      "BluPay",
		"paymentMethod": "pix",
		"status":        "waiting_payment",
		"createdAt":     order.CreatedAt.Format("2006-01-02 15:04:05"),
		"approvedDate":  nil,
		"refundedAt":    nil,
		"customer": map[string]interface{}{
			"name":     customer.Name,
			"email":    customer.Email,
			"phone":    customer.Phone,
			"document": customer.Document,
			"country":  "BR",
			"ip":       customer.IP,
		},
		"products": []map[string]interface{}{
			{
				"id":           "PROD_" + order.TransactionID,
				"name":         s.cfg.BluPayProductName,
				"planId":       nil,
				"planName":     nil,
				"quantity":     1,
				"priceInCents": order.Amount,
			},
		},
		"trackingParameters": trackingParams,
		"commission": map[string]interface{}{
			"totalPriceInCents":     order.Amount,
			"gatewayFeeInCents":     gatewayFee,
			"userCommissionInCents": order.Amount - gatewayFee,
		},
		"isTest": false,
	}

	// Envia para Utmify
	body, _ := json.Marshal(utmifyPayload)
	httpReq, err := http.NewRequest("POST", s.cfg.UtmifyAPIURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("❌ Erro ao criar requisição Utmify: %v", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-token", s.cfg.UtmifyToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("❌ Erro ao enviar para Utmify: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📡 [Utmify Pending - BluPay] HTTP %d: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode == http.StatusOK {
		log.Println("✅ Dados enviados com sucesso para Utmify (pending)")
	} else {
		log.Printf("⚠️ Resposta não-200 do Utmify: %s", string(respBody))
	}
}

// Preenche dados faltantes automaticamente usando 4devs
func (s *BluPayService) fillMissingData(req *dto.BluPayRequest) error {
	// Verifica se precisa gerar dados
	needsFakeData := req.Name == "" || req.Email == "" || req.Document == "" || req.Phone == ""

	if !needsFakeData {
		return nil // Todos os dados fornecidos
	}

	log.Println("🔄 Dados incompletos detectados, gerando automaticamente via 5devs...")

	// Gera pessoa fake
	fakerService := NewFakerService()
	pessoa, err := fakerService.GerarPessoa()
	if err != nil {
		return fmt.Errorf("erro ao gerar dados fake: %w", err)
	}

	// Preenche apenas os campos vazios
	if req.Name == "" {
		req.Name = pessoa.Nome
		log.Printf("✅ Nome gerado: %s", req.Name)
	}

	if req.Email == "" {
		req.Email = pessoa.Email
		log.Printf("✅ Email gerado: %s", req.Email)
	}

	if req.Document == "" {
		req.Document = fakerService.CleanCPF(pessoa.CPF)
		log.Printf("✅ CPF gerado: %s", req.Document)
	}

	if req.Phone == "" {
		req.Phone = fakerService.CleanPhone(pessoa.Celular)
		log.Printf("✅ Telefone gerado: %s", req.Phone)
	}

	return nil
}
