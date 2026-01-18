package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/victtorkaiser/server-apis/internal/config"
	"github.com/victtorkaiser/server-apis/internal/dto"
	"github.com/victtorkaiser/server-apis/internal/models"
	"github.com/victtorkaiser/server-apis/internal/queue"
	"gorm.io/gorm"
)

type WebhookService struct {
	db       *gorm.DB
	redis    *redis.Client
	rabbitMQ *queue.RabbitMQ
	cfg      *config.Config
}

func NewWebhookService(db *gorm.DB, redis *redis.Client, rabbitMQ *queue.RabbitMQ, cfg *config.Config) *WebhookService {
	return &WebhookService{
		db:       db,
		redis:    redis,
		rabbitMQ: rabbitMQ,
		cfg:      cfg,
	}
}

func (s *WebhookService) ProcessWebhook(ctx context.Context, webhook *dto.WebhookPayload) error {
	paymentCode := webhook.GetPaymentCode()
	status := webhook.GetStatus()

	log.Printf("🔄 [Webhook] Processando: Type=%s PaymentCode=%s Status=%s", webhook.Type, paymentCode, status)

	// Valida se tem payment code
	if paymentCode == "" {
		return fmt.Errorf("payment_code/objectId não encontrado no webhook")
	}

	// Busca o pedido
	var order models.Order
	if err := s.db.Preload("Customer").Preload("TrackingParameter").First(&order, "transaction_id = ?", paymentCode).Error; err != nil {
		log.Printf("❌ [Webhook] Pedido não encontrado: transaction_id=%s", paymentCode)
		return fmt.Errorf("pedido não encontrado: %w", err)
	}

	log.Printf("📦 [Webhook] Pedido encontrado: ID=%s Platform=%s OldStatus=%s", order.ID, order.Platform, order.Status)

	// Atualiza o status
	oldStatus := order.Status
	newStatus := s.mapStatus(status)
	order.Status = newStatus
	order.UpdatedAt = time.Now()

	// Se aprovado/pago, marca data de aprovação
	if newStatus == models.OrderStatusApproved || newStatus == models.OrderStatusPaid {
		now := time.Now()
		order.ApprovedAt = &now
		log.Printf("✅ [Webhook] Pagamento aprovado em: %s", now.Format("2006-01-02 15:04:05"))
	}

	// Salva no banco
	if err := s.db.Save(&order).Error; err != nil {
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}

	log.Printf("✅ [Webhook] Status atualizado: %s -> %s (transaction_id=%s)", oldStatus, newStatus, paymentCode)

	// Publica eventos na fila se aprovado (se RabbitMQ estiver disponível)
	if newStatus == models.OrderStatusApproved || newStatus == models.OrderStatusPaid {
		if s.rabbitMQ != nil {
			log.Printf("📤 [Webhook] Publicando eventos payment.approved e utmify.approved")

			s.rabbitMQ.Publish("payment.approved", map[string]interface{}{
				"order_id":       order.ID.String(),
				"transaction_id": order.TransactionID,
				"status":         string(order.Status),
				"platform":       order.Platform,
			})

			// Publica para enviar ao Utmify
			s.rabbitMQ.Publish("utmify.approved", map[string]interface{}{
				"order_id":       order.ID.String(),
				"transaction_id": order.TransactionID,
			})
		} else {
			log.Printf("⚠️ [Webhook] RabbitMQ não disponível, eventos não publicados")
		}

		// Envia webhook para URL externa (assíncrono)
		go s.SendExternalWebhook(&order)
	}

	return nil
}

func (s *WebhookService) mapStatus(status string) models.OrderStatus {
	status = strings.ToUpper(status)
	switch status {
	case "APPROVED", "PAID":
		return models.OrderStatusApproved
	case "PENDING":
		return models.OrderStatusPending
	case "WAITING_PAYMENT":
		return models.OrderStatusWaitingPayment
	case "REFUNDED":
		return models.OrderStatusRefunded
	case "CANCELLED":
		return models.OrderStatusCancelled
	default:
		return models.OrderStatusPending
	}
}

// SendExternalWebhook envia webhook para URL externa do cliente
func (s *WebhookService) SendExternalWebhook(order *models.Order) {
	if order.WebhookURL == "" {
		log.Printf("⚠️ [Webhook Externo] Nenhuma URL configurada para order %s", order.ID)
		return
	}

	log.Printf("📤 [Webhook Externo] Enviando para: %s", order.WebhookURL)

	// Prepara payload do webhook
	payload := map[string]interface{}{
		"event":          "payment.approved",
		"transaction_id": order.TransactionID,
		"order_id":       order.ID.String(),
		"status":         string(order.Status),
		"amount":         order.Amount,
		"payment_method": order.PaymentMethod,
		"platform":       order.Platform,
		"approved_at":    order.ApprovedAt,
		"created_at":     order.CreatedAt,
		"customer": map[string]interface{}{
			"id":       order.Customer.ID.String(),
			"name":     order.Customer.Name,
			"email":    order.Customer.Email,
			"phone":    order.Customer.Phone,
			"document": order.Customer.Document,
		},
	}

	// Se tiver tracking parameters, inclui
	if order.TrackingParameter != nil {
		payload["tracking_params"] = map[string]interface{}{
			"utm_source":   order.TrackingParameter.UtmSource,
			"utm_campaign": order.TrackingParameter.UtmCampaign,
			"utm_medium":   order.TrackingParameter.UtmMedium,
			"utm_content":  order.TrackingParameter.UtmContent,
			"utm_term":     order.TrackingParameter.UtmTerm,
			"gclid":        order.TrackingParameter.Gclid,
			"fbclid":       order.TrackingParameter.Fbclid,
			"ttclid":       order.TrackingParameter.Ttclid,
			"sck":          order.TrackingParameter.Sck,
			"xcod":         order.TrackingParameter.Xcod,
		}
	}

	body, _ := json.Marshal(payload)
	log.Printf("📦 [Webhook Externo] Payload: %s", string(body))

	// Envia webhook com retry (5 tentativas: imediato, 1s, 10s, 30s, 60s)
	maxRetries := 5
	retryIntervals := []time.Duration{
		1 * time.Second,  // Tentativa 2
		10 * time.Second, // Tentativa 3
		30 * time.Second, // Tentativa 4
		60 * time.Second, // Tentativa 5
	}

	for i := 1; i <= maxRetries; i++ {
		httpReq, err := http.NewRequest("POST", order.WebhookURL, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("❌ [Webhook Externo] Erro ao criar requisição: %v", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", "Server-APIs-Webhook/1.0")
		httpReq.Header.Set("X-Webhook-Event", "payment.approved")
		httpReq.Header.Set("X-Transaction-ID", order.TransactionID)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			log.Printf("❌ [Webhook Externo] Tentativa %d/%d falhou: %v", i, maxRetries, err)
			if i < maxRetries {
				interval := retryIntervals[i-1]
				log.Printf("⏳ [Webhook Externo] Aguardando %v antes da próxima tentativa...", interval)
				time.Sleep(interval)
				continue
			}
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("📡 [Webhook Externo] HTTP %d: %s", resp.StatusCode, string(respBody))

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("✅ [Webhook Externo] Enviado com sucesso para %s", order.WebhookURL)
			return
		}

		log.Printf("⚠️ [Webhook Externo] Tentativa %d/%d - HTTP %d", i, maxRetries, resp.StatusCode)
		if i < maxRetries {
			interval := retryIntervals[i-1]
			log.Printf("⏳ [Webhook Externo] Aguardando %v antes da próxima tentativa...", interval)
			time.Sleep(interval)
		}
	}

	log.Printf("❌ [Webhook Externo] Todas as tentativas falharam para %s", order.WebhookURL)
}
