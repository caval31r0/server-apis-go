package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/victtorkaiser/server-apis/internal/config"
	"github.com/victtorkaiser/server-apis/internal/database"
	"github.com/victtorkaiser/server-apis/internal/queue"
	"github.com/victtorkaiser/server-apis/internal/router"
	"github.com/victtorkaiser/server-apis/internal/workers"
)

func main() {
	// Carrega variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Inicializa configuração
	cfg := config.Load()

	// Conecta ao banco de dados
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	// Executa migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Erro ao executar migrations: %v", err)
	}

	// Conecta ao Redis
	redisClient := database.ConnectRedis(cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB)

	// Conecta ao RabbitMQ (opcional)
	var rabbitMQ *queue.RabbitMQ
	if cfg.RabbitMQURL != "" {
		rabbitMQ, err = queue.Connect(cfg.RabbitMQURL)
		if err != nil {
			log.Printf("⚠️ RabbitMQ não conectado: %v (continuando sem filas)", err)
		} else {
			log.Println("✅ RabbitMQ conectado")
			defer rabbitMQ.Close()

			// Inicia consumers
			utmifyConsumer := workers.NewUtmifyConsumer(db, rabbitMQ, cfg)
			if err := utmifyConsumer.Start(); err != nil {
				log.Printf("⚠️ Erro ao iniciar UtmifyConsumer: %v", err)
			}
		}
	} else {
		log.Println("ℹ️ RabbitMQ desabilitado (RABBITMQ_URL vazio)")
	}

	// Configura modo do Gin
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Inicializa router
	r := router.Setup(db, redisClient, rabbitMQ, cfg)

	// Inicia servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Servidor iniciado na porta %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
