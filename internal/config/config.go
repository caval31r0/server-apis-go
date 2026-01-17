package config

import "os"

type Config struct {
	Env           string
	Port          string
	DatabaseURL   string
	RedisURL      string
	RedisPassword string
	RedisDB       int
	RabbitMQURL   string

	// External APIs
	MangoFyAPIURL         string
	MangoFySecret         string
	MangoFyAPIKey         string
	QuantumPayAPIURL      string
	QuantumPaySecretKey   string
	QuantumPayProductName string
	BluPayAPIURL          string
	BluPaySecretKey       string
	BluPayPublicKey       string
	BluPayWebhookSecret   string
	BluPayWebhookURL      string
	BluPayProductName     string
	UtmifyAPIURL          string
	UtmifyToken           string
	WebhookBaseURL        string
	CPFAPIUrl             string
	CPFAPIToken           string
}

func Load() *Config {
	return &Config{
		Env:                 getEnv("ENV", "development"),
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		RedisURL:            getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		RedisDB:             0,
		RabbitMQURL:         getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		MangoFyAPIURL:         getEnv("MANGOFY_API_URL", ""),
		MangoFySecret:         getEnv("MANGOFY_SECRET_KEY", ""),
		MangoFyAPIKey:         getEnv("MANGOFY_API_KEY", ""),
		QuantumPayAPIURL:      getEnv("QUANTUMPAY_API_URL", "https://api.quantumpayments.com.br/v1/transactions"),
		QuantumPaySecretKey:   getEnv("QUANTUMPAY_SECRET_KEY", ""),
		QuantumPayProductName: getEnv("QUANTUMPAY_PRODUCT_NAME", "Produto"),
		BluPayAPIURL:          getEnv("BLUPAY_API_URL", "https://docs.blupayip.io/baseUrl/api/v1"),
		BluPaySecretKey:       getEnv("BLUPAY_SECRET_KEY", "live_-8EI6hKJSkaYUyvyBjBlDZdkfee0hY8_"),
		BluPayPublicKey:       getEnv("BLUPAY_PUBLIC_KEY", "65136884-dd99-4ede-8566-28505082473a"),
		BluPayWebhookSecret:   getEnv("BLUPAY_WEBHOOK_SECRET", "secret_900de97d1cf10dda70c803fede642899"),
		BluPayWebhookURL:      getEnv("BLUPAY_WEBHOOK_URL", ""),
		BluPayProductName:     getEnv("BLUPAY_PRODUCT_NAME", "Produto"),
		UtmifyAPIURL:          getEnv("UTMIFY_API_URL", ""),
		UtmifyToken:           getEnv("UTMIFY_TOKEN", ""),
		WebhookBaseURL:        getEnv("WEBHOOK_BASE_URL", ""),
		CPFAPIUrl:             getEnv("CPF_API_URL", "https://searchapi.dnnl.live/consulta"),
		CPFAPIToken:           getEnv("CPF_API_TOKEN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
