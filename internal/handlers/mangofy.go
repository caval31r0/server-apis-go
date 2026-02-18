package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victtorkaiser/server-apis/internal/dto"
	"github.com/victtorkaiser/server-apis/internal/services"
)

type MangoFyHandler struct {
	service        *services.MangoFyService
	webhookService *services.WebhookService
}

func NewMangoFyHandler(service *services.MangoFyService, webhookService *services.WebhookService) *MangoFyHandler {
	return &MangoFyHandler{
		service:        service,
		webhookService: webhookService,
	}
}

func (h *MangoFyHandler) CreatePayment(c *gin.Context) {
	var req dto.MangoFyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dados inválidos: " + err.Error(),
		})
		return
	}

	resp, err := h.service.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Erro ao criar pagamento: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *MangoFyHandler) HandleWebhook(c *gin.Context) {
	var webhook dto.WebhookPayload
	if err := c.ShouldBindJSON(&webhook); err != nil {
		log.Printf("❌ [MangoFy Webhook] Payload inválido: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	log.Printf("📥 [MangoFy Webhook] Recebido: payment_code=%s status=%s", webhook.PaymentCode, webhook.PaymentStatus)

	if err := h.webhookService.ProcessWebhook(c.Request.Context(), &webhook); err != nil {
		log.Printf("❌ [MangoFy Webhook] Erro ao processar: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
