package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/victtorkaiser/server-apis/internal/config"
	"github.com/victtorkaiser/server-apis/internal/handlers"
	"github.com/victtorkaiser/server-apis/internal/middlewares"
	"github.com/victtorkaiser/server-apis/internal/queue"
	"github.com/victtorkaiser/server-apis/internal/services"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, redis *redis.Client, rabbitMQ *queue.RabbitMQ, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Middlewares globais
	r.Use(middlewares.CORS())
	r.Use(middlewares.Logger())
	r.Use(middlewares.Recovery())

	// Services
	paymentService := services.NewPaymentService(db, redis, rabbitMQ, cfg)
	webhookService := services.NewWebhookService(db, redis, rabbitMQ, cfg)
	utmifyService := services.NewUtmifyService(cfg)
	quantumPayService := services.NewQuantumPayService(db, redis, rabbitMQ, cfg)
	cpfService := services.NewCPFService(cfg)

	// Handlers
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	webhookHandler := handlers.NewWebhookHandler(webhookService, utmifyService)
	healthHandler := handlers.NewHealthHandler(db, redis)
	quantumPayHandler := handlers.NewQuantumPayHandler(quantumPayService)
	cpfHandler := handlers.NewCPFHandler(cpfService)

	// Health check
	r.GET("/health", healthHandler.Check)

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Pagamentos
		payments := v1.Group("/payments")
		{
			payments.POST("", paymentHandler.Create)
			payments.GET("/:id", paymentHandler.GetByID)
			payments.GET("/transaction/:transaction_id", paymentHandler.GetByTransactionID)
		}

		// Webhooks
		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("/payment", webhookHandler.HandlePayment)
		}
	}

	// API Payment (QuantumPay)
	payment := r.Group("/api/payment")
	{
		payment.POST("/quantumpay", quantumPayHandler.CreatePayment)
	}

	// API CPF
	r.GET("/api/cpf/:cpf", cpfHandler.Consultar)
	r.GET("/api/cpf", cpfHandler.Consultar) // Também aceita query string

	return r
}
