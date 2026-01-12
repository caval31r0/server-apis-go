package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victtorkaiser/server-apis/internal/dto"
	"github.com/victtorkaiser/server-apis/internal/services"
)

type WebhookHandler struct {
	webhookService *services.WebhookService
	utmifyService  *services.UtmifyService
}

func NewWebhookHandler(webhookService *services.WebhookService, utmifyService *services.UtmifyService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		utmifyService:  utmifyService,
	}
}

func (h *WebhookHandler) HandlePayment(c *gin.Context) {
	var webhook dto.WebhookPayload
	if err := c.ShouldBindJSON(&webhook); err != nil {
		log.Printf("❌ [Webhook] Payload inválido: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	log.Printf("📥 [Webhook] Recebido: PaymentCode=%s Status=%s", webhook.PaymentCode, webhook.PaymentStatus)

	// Processa o webhook
	if err := h.webhookService.ProcessWebhook(c.Request.Context(), &webhook); err != nil {
		log.Printf("❌ [Webhook] Erro ao processar: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
