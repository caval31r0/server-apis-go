package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/victtorkaiser/server-apis/internal/config"
	"github.com/victtorkaiser/server-apis/internal/dto"
	"github.com/victtorkaiser/server-apis/internal/models"
)

type UtmifyService struct {
	cfg *config.Config
}

func NewUtmifyService(cfg *config.Config) *UtmifyService {
	return &UtmifyService{cfg: cfg}
}

func (s *UtmifyService) SendPendingOrder(order *models.Order) error {
	payload := s.buildUtmifyPayload(order, "waiting_payment")
	return s.sendToUtmify(payload)
}

func (s *UtmifyService) SendApprovedOrder(order *models.Order) error {
	payload := s.buildUtmifyPayload(order, "paid")
	return s.sendToUtmify(payload)
}

func (s *UtmifyService) buildUtmifyPayload(order *models.Order, status string) *dto.UtmifyOrderRequest {
	trackingParams := make(map[string]interface{})
	if order.TrackingParameter != nil {
		tp := order.TrackingParameter
		if tp.Src != "" {
			trackingParams["src"] = tp.Src
		}
		if tp.Sck != "" {
			trackingParams["sck"] = tp.Sck
		}
		if tp.UtmSource != "" {
			trackingParams["utm_source"] = tp.UtmSource
		}
		if tp.UtmCampaign != "" {
			trackingParams["utm_campaign"] = tp.UtmCampaign
		}
		if tp.UtmMedium != "" {
			trackingParams["utm_medium"] = tp.UtmMedium
		}
		if tp.UtmContent != "" {
			trackingParams["utm_content"] = tp.UtmContent
		}
		if tp.UtmTerm != "" {
			trackingParams["utm_term"] = tp.UtmTerm
		}
		if tp.Xcod != "" {
			trackingParams["xcod"] = tp.Xcod
		}
		if tp.Fbclid != "" {
			trackingParams["fbclid"] = tp.Fbclid
		}
		if tp.Gclid != "" {
			trackingParams["gclid"] = tp.Gclid
		}
		if tp.Ttclid != "" {
			trackingParams["ttclid"] = tp.Ttclid
		}
	}

	payload := &dto.UtmifyOrderRequest{
		OrderID:        order.TransactionID,
		Platform:       "PayHubr",
		PaymentMethod:  "pix",
		Status:         status,
		CreatedAt:      order.CreatedAt,
		ApprovedDate:   order.ApprovedAt,
		PaidAt:         order.ApprovedAt,
		RefundedAt:     order.RefundedAt,
		TrackingParams: trackingParams,
		IsTest:         false,
	}

	// Customer
	payload.Customer = dto.UtmifyCustomer{
		Name:     order.Customer.Name,
		Email:    order.Customer.Email,
		Phone:    order.Customer.Phone,
		Document: order.Customer.Document,
		Country:  order.Customer.Country,
		IP:       order.Customer.IP,
	}

	// Products
	if len(order.Products) > 0 {
		for _, p := range order.Products {
			payload.Products = append(payload.Products, dto.UtmifyProduct{
				ID:       p.Code,
				Name:     p.Name,
				PlanID:   p.PlanID,
				PlanName: p.PlanName,
				Quantity: p.Quantity,
				Price:    p.Price,
			})
		}
	} else {
		// Produto padrão se não houver
		payload.Products = []dto.UtmifyProduct{
			{
				ID:       "default",
				Name:     "Produto",
				Quantity: 1,
				Price:    order.Amount,
			},
		}
	}

	// Commission
	payload.Commission = &dto.UtmifyCommission{
		TotalPrice:     order.Amount,
		GatewayFee:     0,
		UserCommission: order.Amount,
	}

	return payload
}

func (s *UtmifyService) sendToUtmify(payload *dto.UtmifyOrderRequest) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	req, err := http.NewRequest("POST", s.cfg.UtmifyAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-token", s.cfg.UtmifyToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📡 [Utmify] Response HTTP %d: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro na API Utmify: status %d - %s", resp.StatusCode, string(respBody))
	}

	log.Printf("✅ [Utmify] Dados enviados com sucesso")
	return nil
}
