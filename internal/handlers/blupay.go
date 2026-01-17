package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victtorkaiser/server-apis/internal/dto"
	"github.com/victtorkaiser/server-apis/internal/services"
)

type BluPayHandler struct {
	service *services.BluPayService
}

func NewBluPayHandler(service *services.BluPayService) *BluPayHandler {
	return &BluPayHandler{
		service: service,
	}
}

func (h *BluPayHandler) CreatePayment(c *gin.Context) {
	var req dto.BluPayRequest
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
