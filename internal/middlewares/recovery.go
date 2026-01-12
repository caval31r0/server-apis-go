package middlewares

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recuperado: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"message": "Ocorreu um erro inesperado",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
