package services

import (
	"context"
	"fmt"
	"log"
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
