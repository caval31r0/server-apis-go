package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/victtorkaiser/server-apis/internal/services"
)

type FreeFireHandler struct {
	service *services.FreeFireService
}

func NewFreeFireHandler(service *services.FreeFireService) *FreeFireHandler {
	return &FreeFireHandler{
		service: service,
	}
}

func (h *FreeFireHandler) GetPlayer(c *gin.Context) {
	playerID := c.Param("id")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Player ID é obrigatório",
		})
		return
	}

	player, err := h.service.GetPlayer(c.Request.Context(), playerID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Erro ao buscar dados do jogador: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, player)
}
