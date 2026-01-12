package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victtorkaiser/server-apis/internal/services"
)

type CPFHandler struct {
	service *services.CPFService
}

func NewCPFHandler(service *services.CPFService) *CPFHandler {
	return &CPFHandler{
		service: service,
	}
}

func (h *CPFHandler) Consultar(c *gin.Context) {
	// Aceita CPF via URL param ou query string
	cpf := c.Param("cpf")
	if cpf == "" {
		cpf = c.Query("cpf")
	}

	if cpf == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "CPF não fornecido",
		})
		return
	}

	result, err := h.service.ConsultarCPF(cpf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
