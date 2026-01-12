package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/victtorkaiser/server-apis/internal/dto"
	"github.com/victtorkaiser/server-apis/internal/services"
)

type PaymentHandler struct {
	service *services.PaymentService
}

func NewPaymentHandler(service *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) Create(c *gin.Context) {
	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos", "details": err.Error()})
		return
	}

	response, err := h.service.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar pagamento", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PaymentHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pedido não encontrado"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *PaymentHandler) GetByTransactionID(c *gin.Context) {
	transactionID := c.Param("transaction_id")

	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID da transação não fornecido",
		})
		return
	}

	order, err := h.service.GetOrderByTransactionID(c.Request.Context(), transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Transação não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"transaction_id": order.TransactionID,
		"status":         order.Status,
		"valor":          order.Amount,
		"platform":       order.Platform,
		"pix_code":       order.PixCode,
		"created_at":     order.CreatedAt,
		"updated_at":     order.UpdatedAt,
		"approved_at":    order.ApprovedAt,
		"nome":           order.Customer.Name,
		"email":          order.Customer.Email,
		"cpf":            order.Customer.Document,
		"data": gin.H{
			"amount":       order.Amount,
			"payment_method": order.PaymentMethod,
			"platform":     order.Platform,
			"created_at":   order.CreatedAt,
			"updated_at":   order.UpdatedAt,
			"approved_at":  order.ApprovedAt,
			"customer": gin.H{
				"name":     order.Customer.Name,
				"email":    order.Customer.Email,
				"document": order.Customer.Document,
				"phone":    order.Customer.Phone,
				"country":  order.Customer.Country,
			},
		},
	})
}
