package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

type GenesysService struct {
	db       *gorm.DB
	redis    *redis.Client
	rabbitMQ *queue.RabbitMQ
	cfg      *config.Config
}

func NewGenesysService(db *gorm.DB, redis *redis.Client, rabbitMQ *queue.RabbitMQ, cfg *config.Config) *GenesysService {
	return &GenesysService{
		db:       db,
		redis:    redis,
		rabbitMQ: rabbitMQ,
		cfg:      cfg,
	}
}

func (s *GenesysService) CreatePayment(ctx context.Context, req *dto.GenesysRequest) (*dto.GenesysResponse, error) {
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

	// Chama API Genesys
	externalResp, err := s.callGenesysAPI(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar API Genesys: %w", err)
	}

	pixCode := externalResp.Pix.Payload

	// Cria order
	order := &models.Order{
		TransactionID:       externalResp.ID,
		Status:              models.OrderStatusPending,
		Amount:              req.Amount,
		PaymentMethod:       "pix",
		Platform:            "Genesys",
		PixCode:             pixCode,
		WebhookURL:          req.WebhookURL,
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
			"platform":       "Genesys",
		})
	}

	return &dto.GenesysResponse{
		Success:   true,
		Token:     order.TransactionID,
		PixCode:   pixCode,
		QRCodeURL: s.generateQRCodeURL(pixCode),
		Amount:    order.Amount,
		Nome:      customer.Name,
		CPF:       customer.Document,
	}, nil
}

func (s *GenesysService) callGenesysAPI(req *dto.GenesysRequest) (*dto.GenesysAPIResponse, error) {
	externalID := fmt.Sprintf("order_%s", uuid.New().String())

	// Converte centavos para BRL (Genesys usa float em reais)
	amountBRL := float64(req.Amount) / 100.0

	payload := dto.GenesysAPIRequest{
		ExternalID:    externalID,
		TotalAmount:   amountBRL,
		PaymentMethod: "PIX",
		WebhookURL:    fmt.Sprintf("%s/api/v1/webhooks/genesys", s.cfg.WebhookBaseURL),
		Items: []dto.GenesysItem{
			{
				ID:          "1",
				Title:       "Produto",
				Description: "Produto digital",
				Price:       amountBRL,
				Quantity:    1,
				IsPhysical:  false,
			},
		},
		IP: func() string {
			if req.IP != "" {
				return req.IP
			}
			return "177.0.0.1"
		}(),
		Customer: dto.GenesysCustomer{
			Name:         req.Name,
			Email:        req.Email,
			Phone:        req.Phone,
			DocumentType: "CPF",
			Document:     req.Document,
		},
	}

	if req.IP != "" {
		payload.IP = req.IP
	}

	body, _ := json.Marshal(payload)
	log.Printf("📤 [Genesys] Request: %s", string(body))

	httpReq, err := http.NewRequest("POST", s.cfg.GenesysAPIURL+"/v1/transactions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("api-secret", s.cfg.GenesysAPISecret)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📡 [Genesys] Response HTTP %d: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("erro na API Genesys: status %d - %s", resp.StatusCode, string(respBody))
	}

	var result dto.GenesysAPIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta Genesys: %w", err)
	}

	if result.HasError {
		return nil, fmt.Errorf("erro retornado pela API Genesys")
	}

	if result.ID == "" {
		return nil, fmt.Errorf("ID não encontrado na resposta da API Genesys")
	}

	return &result, nil
}

func (s *GenesysService) generateQRCodeURL(pixCode string) string {
	if pixCode != "" {
		return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s", pixCode)
	}
	return ""
}

func (s *GenesysService) mapTrackingParams(params map[string]interface{}) models.TrackingParameter {
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

func (s *GenesysService) fillMissingData(req *dto.GenesysRequest) error {
	needsFakeData := req.Name == "" || req.Email == "" || req.Document == "" || req.Phone == ""

	if !needsFakeData {
		return nil
	}

	log.Println("🔄 Dados incompletos detectados, gerando automaticamente via 5devs...")

	fakerService := NewFakerService()
	pessoa, err := fakerService.GerarPessoa()
	if err != nil {
		return fmt.Errorf("erro ao gerar dados fake: %w", err)
	}

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
