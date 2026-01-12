package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victtorkaiser/server-apis/internal/dto"
	"github.com/victtorkaiser/server-apis/internal/services"
)

type QuantumPayHandler struct {
	service *services.QuantumPayService
}

func NewQuantumPayHandler(service *services.QuantumPayService) *QuantumPayHandler {
	return &QuantumPayHandler{
		service: service,
	}
}

func (h *QuantumPayHandler) CreatePayment(c *gin.Context) {
	var req dto.QuantumPayRequest
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
