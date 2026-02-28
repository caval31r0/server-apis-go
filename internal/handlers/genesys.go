package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victtorkaiser/server-apis/internal/dto"
	"github.com/victtorkaiser/server-apis/internal/services"
)

type GenesysHandler struct {
	service        *services.GenesysService
	webhookService *services.WebhookService
}

func NewGenesysHandler(service *services.GenesysService, webhookService *services.WebhookService) *GenesysHandler {
	return &GenesysHandler{
		service:        service,
		webhookService: webhookService,
	}
}

func (h *GenesysHandler) CreatePayment(c *gin.Context) {
	var req dto.GenesysRequest
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

func (h *GenesysHandler) HandleWebhook(c *gin.Context) {
	var webhook dto.GenesysWebhookPayload
	if err := c.ShouldBindJSON(&webhook); err != nil {
		log.Printf("❌ [Genesys Webhook] Payload inválido: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	log.Printf("📥 [Genesys Webhook] Recebido: id=%s status=%s amount=%.2f", webhook.ID, webhook.Status, webhook.TotalAmount)

	// Converte para o formato genérico do WebhookPayload
	genericWebhook := &dto.WebhookPayload{
		PaymentCode:   webhook.ID,
		PaymentStatus: webhook.Status,
	}

	if err := h.webhookService.ProcessWebhook(c.Request.Context(), genericWebhook); err != nil {
		log.Printf("❌ [Genesys Webhook] Erro ao processar: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
