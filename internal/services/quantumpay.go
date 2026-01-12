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

type QuantumPayService struct {
	db       *gorm.DB
	redis    *redis.Client
	rabbitMQ *queue.RabbitMQ
	cfg      *config.Config
}

func NewQuantumPayService(db *gorm.DB, redis *redis.Client, rabbitMQ *queue.RabbitMQ, cfg *config.Config) *QuantumPayService {
	return &QuantumPayService{
		db:       db,
		redis:    redis,
		rabbitMQ: rabbitMQ,
		cfg:      cfg,
	}
}

func (s *QuantumPayService) CreatePayment(ctx context.Context, req *dto.QuantumPayRequest) (*dto.QuantumPayResponse, error) {
	// Gera placa aleatória para referência externa
	placa := s.generatePlaca()

	// Gera dados automaticamente se não fornecidos
	if err := s.fillMissingData(req); err != nil {
		log.Printf("⚠️ Erro ao gerar dados automáticos: %v (continuando com dados fornecidos)", err)
	}

	// Cria customer
	customer := &models.Customer{
		Name:     req.Name,
		Email:    req.Email,
		Document: req.Document,
		Phone:    req.Telephone,
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

	// Chama API QuantumPay
	externalResp, err := s.callQuantumPayAPI(req, placa)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar API QuantumPay: %w", err)
	}

	// Extrai dados do PIX
	pixCode := s.extractPixCode(externalResp)
	qrCodeURL := s.extractQRCodeURL(externalResp, pixCode)
	txid := s.extractTxid(externalResp)
	transactionID := s.extractTransactionID(externalResp)

	// Cria order
	order := &models.Order{
		TransactionID:       transactionID,
		Status:              models.OrderStatusPending,
		Amount:              req.Amount,
		PaymentMethod:       "pix",
		Platform:            "QuantumPay",
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
			"platform":       "QuantumPay",
		})
	}

	// Envia para Utmify (pendente)
	go s.sendToUtmifyPending(order, customer, req, trackingParamID, externalResp.Fee.Amount)

	return &dto.QuantumPayResponse{
		Success:   true,
		Token:     order.TransactionID,
		PixCode:   pixCode,
		QRCodeURL: qrCodeURL,
		Amount:    order.Amount,
		Nome:      customer.Name,
		CPF:       customer.Document,
		ExpiraEm:  "1 dia",
		Txid:      txid,
	}, nil
}

func (s *QuantumPayService) callQuantumPayAPI(req *dto.QuantumPayRequest, placa string) (*dto.QuantumPayAPIResponse, error) {
	// Monta metadata com UTM params
	metadataJSON, _ := json.Marshal(req.UTMParams)

	// Prepara payload conforme API QuantumPay
	payload := dto.QuantumPayAPIRequest{
		Amount:        req.Amount,
		PaymentMethod: "pix",
		Pix: dto.QuantumPayPixConfig{
			ExpiresInDays: 1,
		},
		Customer: dto.QuantumPayCustomer{
			Name:  req.Name,
			Email: req.Email,
			Phone: req.Telephone,
			Document: dto.QuantumPayDocument{
				Type:   "cpf",
				Number: req.Document,
			},
			ExternalRef: fmt.Sprintf("md-%s-%d", placa, time.Now().Unix()),
		},
		Items: []dto.QuantumPayItem{
			{
				Title:       s.cfg.QuantumPayProductName,
				UnitPrice:   req.Amount,
				Quantity:    1,
				Tangible:    false,
				ExternalRef: fmt.Sprintf("IPVA-%s", placa),
			},
		},
		Metadata: string(metadataJSON),
		IP:       "127.0.0.1",
	}

	body, _ := json.Marshal(payload)
	log.Printf("📤 [QuantumPay] Request: %s", string(body))

	httpReq, err := http.NewRequest("POST", s.cfg.QuantumPayAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Autenticação Basic com secret key
	auth := base64.StdEncoding.EncodeToString([]byte(s.cfg.QuantumPaySecretKey + ":x"))
	httpReq.Header.Set("Authorization", "Basic "+auth)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📡 [QuantumPay] Response: %s", string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("erro na API QuantumPay: status %d - %s", resp.StatusCode, string(respBody))
	}

	var result dto.QuantumPayAPIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.ID == nil {
		return nil, fmt.Errorf("ID não encontrado na resposta da API QuantumPay")
	}

	return &result, nil
}

func (s *QuantumPayService) extractTransactionID(resp *dto.QuantumPayAPIResponse) string {
	// Converte ID (pode ser string ou número) para string
	switch v := resp.ID.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (s *QuantumPayService) extractPixCode(resp *dto.QuantumPayAPIResponse) string {
	if resp.Pix.QRCode != "" {
		return resp.Pix.QRCode
	}
	return ""
}

func (s *QuantumPayService) extractQRCodeURL(resp *dto.QuantumPayAPIResponse, pixCode string) string {
	// Tenta usar URL da API primeiro
	if resp.Pix.ReceiptURL != "" {
		return resp.Pix.ReceiptURL
	}
	if resp.Pix.QRCodeURL != "" {
		return resp.Pix.QRCodeURL
	}

	// Gera QR Code usando QRServer como fallback
	if pixCode != "" {
		return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s", pixCode)
	}

	return ""
}

func (s *QuantumPayService) extractTxid(resp *dto.QuantumPayAPIResponse) string {
	if resp.Pix.End2EndID != "" {
		return resp.Pix.End2EndID
	}
	if resp.Pix.Txid != "" {
		return resp.Pix.Txid
	}
	return ""
}

func (s *QuantumPayService) generatePlaca() string {
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

func (s *QuantumPayService) mapTrackingParams(params map[string]interface{}) models.TrackingParameter {
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

func (s *QuantumPayService) sendToUtmifyPending(order *models.Order, customer *models.Customer, req *dto.QuantumPayRequest, trackingParamID *uuid.UUID, gatewayFee int) {
	// Publica na fila para processar de forma assíncrona (se RabbitMQ estiver disponível)
	if s.rabbitMQ != nil {
		s.rabbitMQ.Publish("utmify.pending", map[string]interface{}{
			"order_id":              order.ID,
			"transaction_id":        order.TransactionID,
			"customer_id":           customer.ID,
			"tracking_parameter_id": trackingParamID,
			"platform":              "QuantumPay",
			"gateway_fee":           gatewayFee,
		})
	}

	// Envia diretamente para Utmify (como no PHP original)
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
		"platform":      "QuantumPay",
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
				"name":         s.cfg.QuantumPayProductName,
				"planId":       nil,
				"planName":     nil,
				"quantity":     1,
				"priceInCents": order.Amount,
			},
		},
		"trackingParameters": trackingParams,
		"commission": map[string]interface{}{
			"totalPriceInCents":      order.Amount,
			"gatewayFeeInCents":      gatewayFee,
			"userCommissionInCents":  order.Amount - gatewayFee,
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
	log.Printf("📡 [Utmify Pending] HTTP %d: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode == http.StatusOK {
		log.Println("✅ Dados enviados com sucesso para Utmify (pending)")
	} else {
		log.Printf("⚠️ Resposta não-200 do Utmify: %s", string(respBody))
	}
}

// Helper para extrair string de map ou retornar nil
func getStringOrNull(params map[string]interface{}, key string) interface{} {
	if val, ok := params[key]; ok && val != nil {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	return nil
}

// Preenche dados faltantes automaticamente usando 4devs
func (s *QuantumPayService) fillMissingData(req *dto.QuantumPayRequest) error {
	// Verifica se precisa gerar dados
	needsFakeData := req.Name == "" || req.Email == "" || req.Document == "" || req.Telephone == ""

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

	if req.Telephone == "" {
		req.Telephone = fakerService.CleanPhone(pessoa.Celular)
		log.Printf("✅ Telefone gerado: %s", req.Telephone)
	}

	return nil
}
