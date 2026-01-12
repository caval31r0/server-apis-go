# Backend APIs - Go

Arquitetura backend profissional em Go com PostgreSQL, Redis e RabbitMQ para gerenciamento de pagamentos e integrações.

## 🏗️ Arquitetura

```
├── cmd/
│   └── api/              # Entry point da aplicação
├── internal/
│   ├── config/           # Configurações
│   ├── database/         # Conexões com bancos
│   ├── dto/              # Data Transfer Objects
│   ├── handlers/         # HTTP handlers
│   ├── middlewares/      # Middlewares HTTP
│   ├── models/           # Modelos do banco
│   ├── queue/            # RabbitMQ
│   ├── router/           # Configuração de rotas
│   └── services/         # Lógica de negócio
└── EXEMPLOS/             # APIs PHP originais (referência)
```

## 🚀 Tecnologias

- **Go 1.21+**
- **PostgreSQL** - Banco de dados principal
- **Redis** - Cache e sessões
- **RabbitMQ** - Filas de mensagens
- **Gin** - Framework HTTP
- **GORM** - ORM

## 📋 Funcionalidades

### APIs Implementadas

1. **POST /api/v1/payments** - Cria pagamento PIX
2. **GET /api/v1/payments/:id** - Busca pedido por ID
3. **GET /api/v1/payments/transaction/:transaction_id** - Busca por transaction_id
4. **POST /api/v1/webhooks/payment** - Recebe webhooks de pagamento
5. **GET /health** - Health check

### Integrações

- **MangoFy** - Gateway de pagamento
- **Utmify** - Tracking de conversões

## ⚙️ Configuração

1. Clone o repositório
2. Configure as variáveis de ambiente (copie `.env.example` para `.env`)
3. Suba os containers:
```bash
make docker-up
```

4. Execute a aplicação:
```bash
make run
```

## 🔧 Variáveis de Ambiente

```env
DATABASE_URL=postgresql://user:pass@host:5432/dbname
REDIS_URL=localhost:6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
MANGOFY_SECRET_KEY=seu_secret_key
MANGOFY_API_KEY=seu_api_key
UTMIFY_TOKEN=seu_token
WEBHOOK_BASE_URL=https://seudominio.com
```

## 🔄 Fluxo de Pagamento

1. Cliente cria pagamento via **POST /api/v1/payments**
2. Sistema chama MangoFy e gera PIX
3. Salva pedido no PostgreSQL
4. Publica evento `payment.created` no RabbitMQ
5. Envia ordem pendente para Utmify
6. Webhook recebe notificação de pagamento aprovado
7. Atualiza status no banco
8. Publica evento `payment.approved`
9. Envia ordem aprovada para Utmify

## 📊 Filas RabbitMQ

- `payment.created` - Pagamento criado
- `payment.approved` - Pagamento aprovado
- `utmify.pending` - Enviar para Utmify (pendente)
- `utmify.approved` - Enviar para Utmify (aprovado)

## 🛠️ Comandos

```bash
make run         # Executa aplicação
make build       # Compila
make test        # Testes
make docker-up   # Sobe containers
make docker-down # Para containers
```

## 📝 Exemplo de Request

```bash
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 2790,
    "name": "João Silva",
    "email": "joao@example.com",
    "document": "12345678900",
    "telephone": "11999999999",
    "utm_params": {
      "utm_source": "google",
      "utm_campaign": "black_friday"
    }
  }'
```

## 📦 Deploy

### Railway (Recomendado)

1. Conecte seu repositório
2. Configure as variáveis de ambiente
3. Railway detectará automaticamente o Go
4. Deploy automático

### Docker

```bash
docker build -t server-apis .
docker run -p 8080:8080 --env-file .env server-apis
```
