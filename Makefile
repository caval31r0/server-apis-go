.PHONY: help run build test clean docker-up docker-down migrate

help:
	@echo "Comandos disponíveis:"
	@echo "  make run         - Executa a aplicação"
	@echo "  make build       - Compila a aplicação"
	@echo "  make test        - Executa os testes"
	@echo "  make clean       - Remove arquivos compilados"
	@echo "  make docker-up   - Sobe containers Docker"
	@echo "  make docker-down - Para containers Docker"
	@echo "  make migrate     - Executa migrations do banco"

run:
	go run cmd/api/main.go

build:
	go build -o bin/server cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	@echo "Migrations são executadas automaticamente ao iniciar a aplicação"
