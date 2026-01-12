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

type PaymentService struct {
	db       *gorm.DB
	redis    *redis.Client
	rabbitMQ *queue.RabbitMQ
	cfg      *config.Config
}

func NewPaymentService(db *gorm.DB, redis *redis.Client, rabbitMQ *queue.RabbitMQ, cfg *config.Config) *PaymentService {
	return &PaymentService{
		db:       db,
		redis:    redis,
		rabbitMQ: rabbitMQ,
		cfg:      cfg,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.CreatePaymentResponse, error) {
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

	// Chama API externa (MangoFy)
	externalResp, err := s.callMangoFyAPI(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar API externa: %w", err)
	}

	// Cria order
	order := &models.Order{
		TransactionID:       externalResp.PaymentCode,
		Status:              models.OrderStatusPending,
		Amount:              req.Amount,
		PaymentMethod:       "pix",
		Platform:            "PayHubr",
		PixCode:             externalResp.PixCode,
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
		})
	}

	// Envia para Utmify (pendente)
	go s.sendToUtmifyPending(order, customer, trackingParamID)

	return &dto.CreatePaymentResponse{
		Success:   true,
		Token:     order.TransactionID,
		PixCode:   order.PixCode,
		QRCodeURL: s.generateQRCodeURL(order.PixCode),
		Amount:    order.Amount,
	}, nil
}

func (s *PaymentService) GetOrderByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	if err := s.db.Preload("Customer").Preload("TrackingParameter").First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *PaymentService) GetOrderByTransactionID(ctx context.Context, transactionID string) (*models.Order, error) {
	var order models.Order
	if err := s.db.Preload("Customer").Preload("TrackingParameter").First(&order, "transaction_id = ?", transactionID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *PaymentService) callMangoFyAPI(req *dto.CreatePaymentRequest) (*mangoFyResponse, error) {
	payload := map[string]interface{}{
		"store_code":      s.cfg.MangoFyAPIKey,
		"external_code":   fmt.Sprintf("order_%s", uuid.New().String()),
		"payment_method":  "pix",
		"payment_format":  "regular",
		"installments":    1,
		"payment_amount":  req.Amount,
		"shipping_amount": 0,
		"postback_url":    fmt.Sprintf("%s/api/v1/webhooks/payment", s.cfg.WebhookBaseURL),
		"items": []map[string]interface{}{
			{
				"code":   "1",
				"name":   "Produto",
				"amount": req.Amount,
				"total":  1,
			},
		},
		"customer": map[string]interface{}{
			"email":    req.Email,
			"name":     req.Name,
			"document": req.Document,
			"phone":    req.Telephone,
		},
		"pix": map[string]interface{}{
			"expires_in_days": 1,
		},
		"extra": req.UTMParams,
	}

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest("POST", s.cfg.MangoFyAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", s.cfg.MangoFySecret)
	httpReq.Header.Set("Store-Code", s.cfg.MangoFyAPIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📡 [MangoFy] Response: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na API MangoFy: status %d", resp.StatusCode)
	}

	var result mangoFyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *PaymentService) mapTrackingParams(params map[string]interface{}) models.TrackingParameter {
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

func (s *PaymentService) generateQRCodeURL(pixCode string) string {
	if pixCode == "" {
		return ""
	}
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?data=%s&size=300x300", pixCode)
}

func (s *PaymentService) sendToUtmifyPending(order *models.Order, customer *models.Customer, trackingParamID *uuid.UUID) {
	// Publica na fila para processar de forma assíncrona (se RabbitMQ estiver disponível)
	if s.rabbitMQ != nil {
		s.rabbitMQ.Publish("utmify.pending", map[string]interface{}{
			"order_id":             order.ID,
			"transaction_id":       order.TransactionID,
			"customer_id":          customer.ID,
			"tracking_parameter_id": trackingParamID,
		})
	}
}

type mangoFyResponse struct {
	PaymentCode string `json:"payment_code"`
	PixCode     string `json:"pix_code"`
	Pix         struct {
		PixQRCodeText string `json:"pix_qrcode_text"`
		PixLink       string `json:"pix_link"`
	} `json:"pix"`
}
