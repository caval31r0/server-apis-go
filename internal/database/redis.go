package database

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(url, password string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       db,
	})

	// Testa a conexão
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		panic("Erro ao conectar ao Redis: " + err.Error())
	}

	return client
}
