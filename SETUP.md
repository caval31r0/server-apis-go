# 🚀 Setup do Projeto

## 1️⃣ Instalação

### Pré-requisitos
- Go 1.21+
- Docker e Docker Compose
- Git

### Clone e Configure

```bash
cd /Users/victtorkaiser/Downloads/SERVER-APIS

# Instale as dependências Go
go mod download

# Configure o ambiente
cp .env.example .env
```

## 2️⃣ Configuração do Banco de Dados

### Railway PostgreSQL (Produção)

Edite o arquivo `.env` e adicione suas credenciais do Railway:

```env
# Server
PORT=8080
ENV=production

# Database - Railway PostgreSQL
DATABASE_URL=postgresql://postgres:********@centerbeam.proxy.rlwy.net:56690/railway

# Redis - Railway
REDIS_URL=maglev.proxy.rlwy.net:45565
REDIS_PASSWORD=IexGkPGDpydlQdXJXcLQxvMCxYHXBRRZ
REDIS_DB=0

# RabbitMQ (local via Docker - opcional)
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# External APIs
MANGOFY_API_URL=https://checkout.mangofy.com.br/api/v1/payment
MANGOFY_SECRET_KEY=2a43ff5154a001bce29e0c749d3f583b4cdtmcbxpze7btvusxg5cxtlxn8zuwq
MANGOFY_API_KEY=3651e35a1fe072e1c4fb19bc54e7ac70

UTMIFY_API_URL=https://api.utmify.com.br/api-credentials/orders
UTMIFY_TOKEN=seu_token_aqui

# Webhook
WEBHOOK_BASE_URL=https://seudominio.com
```

### Teste a Conexão com o Banco

```bash
# Via psql (se tiver instalado)
PGPASSWORD=******** psql -h centerbeam.proxy.rlwy.net -U postgres -p 56690 -d railway

# OU via Railway CLI
railway connect Postgres
```

## 3️⃣ Inicie os Serviços Locais (Opcional)

**Nota**: PostgreSQL e Redis já estão no Railway. RabbitMQ é opcional.

```bash
# Se quiser usar RabbitMQ local (opcional)
make docker-up

# Verifique se está rodando
docker ps
```

Você verá:
- **RabbitMQ** em `localhost:5672` (Management UI em `http://localhost:15672`)

**Usando apenas Railway (sem Docker local)**:
- PostgreSQL: `centerbeam.proxy.rlwy.net:56690`
- Redis: `maglev.proxy.rlwy.net:45565`
- RabbitMQ: Opcional (pode pular filas por enquanto)

## 4️⃣ Execute a Aplicação

```bash
# Modo desenvolvimento
make run

# OU compile e execute
make build
./bin/server
```

A aplicação iniciará em `http://localhost:8080`

## 5️⃣ Teste as APIs

### Health Check
```bash
curl http://localhost:8080/health
```

### Criar Pagamento
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
      "utm_campaign": "teste"
    }
  }'
```

### Consultar Pagamento
```bash
# Por transaction_id
curl http://localhost:8080/api/v1/payments/transaction/SEU_TRANSACTION_ID
```

## 6️⃣ RabbitMQ Management

Acesse: `http://localhost:15672`
- **User**: guest
- **Password**: guest

Você verá as filas:
- `payment.created`
- `payment.approved`
- `utmify.pending`
- `utmify.approved`

## 7️⃣ Migrations

As migrations são executadas automaticamente ao iniciar a aplicação via GORM AutoMigrate.

Tabelas criadas:
- `orders` - Pedidos
- `customers` - Clientes
- `products` - Produtos
- `tracking_parameters` - Parâmetros UTM
- `order_products` - Relacionamento N:N

## 8️⃣ Comandos Úteis

```bash
make run         # Executa aplicação
make build       # Compila
make test        # Testes
make clean       # Limpa binários
make docker-up   # Sobe containers locais
make docker-down # Para containers
```

## 🔥 Deploy (Railway ou Outro)

### Via Railway CLI

```bash
# Instale o Railway CLI
npm i -g @railway/cli

# Login
railway login

# Link ao projeto
railway link

# Deploy
railway up
```

### Via Docker

```bash
docker build -t server-apis .
docker run -p 8080:8080 --env-file .env server-apis
```

## 📊 Monitoramento

### Logs da Aplicação
```bash
# Durante desenvolvimento
make run

# Em produção, use logs do Railway
railway logs
```

### PostgreSQL (Railway)
```bash
# Conecte via Railway CLI
railway connect Postgres

# Liste as tabelas
\dt

# Consulte pedidos
SELECT * FROM orders LIMIT 10;
```

## ⚠️ Custos do Railway

> **Atenção**: Conectar via rede pública causa custos de Egress no Railway.

Para economizar:
- Use a Railway CLI para queries manuais
- Minimize consultas diretas ao banco
- Use Redis para cache quando possível
- Configure limits no Railway

## 🎯 Próximos Passos

1. ✅ Configure as credenciais do Utmify no `.env`
2. ✅ Configure o `WEBHOOK_BASE_URL` (domínio onde a API estará hospedada)
3. ✅ Teste o fluxo completo de pagamento
4. ✅ Configure webhook no MangoFy apontando para seu domínio
5. ✅ Implemente workers para consumir filas RabbitMQ (opcional)

## 📝 Estrutura de Pastas

```
SERVER-APIS/
├── cmd/api/                # Entry point
├── internal/
│   ├── config/             # Configs centralizadas
│   ├── database/           # PostgreSQL + Redis
│   ├── dto/                # Request/Response objects
│   ├── handlers/           # HTTP handlers (controllers)
│   ├── middlewares/        # CORS, Logger, Recovery
│   ├── models/             # Database models
│   ├── queue/              # RabbitMQ
│   ├── router/             # Routes setup
│   └── services/           # Business logic
├── EXEMPLOS/               # APIs PHP (referência)
├── .env                    # Configurações (não versionar)
├── docker-compose.yml      # Redis + RabbitMQ
├── Dockerfile              # Build da app
├── Makefile                # Comandos úteis
└── README.md               # Documentação
```

## 🆘 Troubleshooting

### Erro: "dial tcp: connection refused" (PostgreSQL)
- Verifique se a URL do Railway está correta no `.env`
- Teste a conexão com `railway connect Postgres`

### Erro: "dial tcp: connection refused" (Redis)
- Execute `make docker-up` para subir o Redis local
- Verifique com `docker ps`

### Erro: "dial tcp: connection refused" (RabbitMQ)
- Execute `make docker-up` para subir o RabbitMQ
- Acesse `http://localhost:15672` para verificar

### Migrations não executam
- Verifique os logs ao iniciar a aplicação
- As migrations são automáticas via GORM AutoMigrate
