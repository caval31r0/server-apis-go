# API Documentation - Server APIs Go

**Base URL:** `https://server-apis-go-production.up.railway.app`

Servidor backend completo para processamento de pagamentos PIX com múltiplos gateways (QuantumPay e BluPay), integração com Utmify para rastreamento de conversões, consulta de CPF e geração automática de dados de teste.

---

## Índice

1. [Health Check](#health-check)
2. [Pagamentos](#pagamentos)
   - [QuantumPay](#quantumpay)
   - [BluPay](#blupay)
   - [Consultar Pagamento](#consultar-pagamento)
3. [Webhooks](#webhooks)
   - [Receber Webhooks](#receber-webhooks)
   - [Webhooks Externos](#webhooks-externos)
4. [Consulta CPF](#consulta-cpf)
5. [Códigos de Status](#códigos-de-status)
6. [Erros Comuns](#erros-comuns)

---

## Health Check

### GET `/health`

Verifica o status da API e conexões com banco de dados e Redis.

#### Response (200 OK)

```json
{
  "status": "ok",
  "database": "connected",
  "redis": "connected",
  "timestamp": "2026-01-17T06:11:53Z"
}
```

#### cURL

```bash
curl https://server-apis-go-production.up.railway.app/health
```

---

## Pagamentos

### QuantumPay

#### POST `/api/payment/quantumpay`

Cria uma cobrança PIX através do gateway QuantumPay.

**Request Body:**

```json
{
  "amount": 10000,
  "name": "João Silva",
  "email": "joao@email.com",
  "document": "12345678910",
  "telephone": "11999999999",
  "utm_params": {
    "utm_source": "google",
    "utm_medium": "cpc",
    "utm_campaign": "black-friday",
    "utm_content": "banner-1",
    "utm_term": "pix",
    "sck": "abc123",
    "xcod": "xyz789",
    "fbclid": "fb123",
    "gclid": "gl123",
    "ttclid": "tt123"
  }
}
```

**Campos:**
- `amount` (obrigatório): Valor em centavos (mínimo: 100 = R$ 1,00)
- `name` (opcional): Nome do cliente (auto-gerado se vazio)
- `email` (opcional): Email do cliente (auto-gerado se vazio)
- `document` (opcional): CPF do cliente sem formatação (auto-gerado se vazio)
- `telephone` (opcional): Telefone com DDD (auto-gerado se vazio)
- `utm_params` (opcional): Parâmetros de rastreamento UTM

**Response (200 OK):**

```json
{
  "success": true,
  "token": "123456789",
  "pixCode": "00020101021226880014br.gov.bcb.pix...",
  "qrCodeUrl": "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=...",
  "amount": 10000,
  "nome": "João Silva",
  "cpf": "12345678910",
  "expiraEm": "1 dia",
  "txid": "E18236120202512020455s14af098224"
}
```

**cURL:**

```bash
curl -X POST https://server-apis-go-production.up.railway.app/api/payment/quantumpay \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 10000,
    "name": "João Silva",
    "email": "joao@email.com",
    "document": "12345678910",
    "telephone": "11999999999",
    "utm_params": {
      "utm_source": "google",
      "utm_campaign": "black-friday",
      "gclid": "abc123"
    }
  }'
```

---

### BluPay

#### POST `/api/payment/blupay`

Cria uma cobrança PIX através do gateway BluPay.

**Request Body:**

```json
{
  "amount": 10000,
  "name": "Maria Santos",
  "email": "maria@email.com",
  "document": "98765432100",
  "phone": "11988887777",
  "externalRef": "ORD-123",
  "utm_params": {
    "utm_source": "google",
    "utm_medium": "cpc",
    "utm_campaign": "23455227240",
    "utm_content": "792714546957",
    "utm_term": "b",
    "keyword": "poupatempo",
    "device": "m",
    "network": "g",
    "matchtype": "b",
    "gad_source": "1",
    "gad_campaignid": "23455227240",
    "gbraid": "0AAAABCYEGMGzBNrjYICxQ9sUoQhJDcI5n",
    "gclid": "CjwKCAiA4KfLBhB0EiwAUY7GAaMKjCUR8dR..."
  }
}
```

**Campos:**
- `amount` (obrigatório): Valor em centavos (mínimo: 100 = R$ 1,00)
- `name` (opcional): Nome do cliente (auto-gerado se vazio)
- `email` (opcional): Email do cliente (auto-gerado se vazio)
- `document` (opcional): CPF sem formatação (auto-gerado se vazio)
- `phone` (opcional): Telefone com DDD (auto-gerado se vazio)
- `externalRef` (opcional): Referência externa (auto-gerado se vazio)
- `utm_params` (opcional): Parâmetros de rastreamento UTM + extras Google Ads

**Response (200 OK):**

```json
{
  "success": true,
  "token": "d22ed312-0959-45df-b09a-c4c9780b62a2",
  "pixCode": "00020101021226850014br.gov.bcb.pix...",
  "qrCodeUrl": "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=...",
  "amount": 10000,
  "nome": "Maria Santos",
  "cpf": "98765432100",
  "expiraEm": "2 dias"
}
```

**cURL (Minimal):**

```bash
curl -X POST https://server-apis-go-production.up.railway.app/api/payment/blupay \
  -H "Content-Type: application/json" \
  -d '{"amount": 10000}'
```

**cURL (Completo com UTMs):**

```bash
curl -X POST https://server-apis-go-production.up.railway.app/api/payment/blupay \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 10000,
    "utm_params": {
      "utm_source": "google",
      "utm_campaign": "23455227240",
      "utm_medium": "193094442473",
      "utm_content": "792714546957",
      "utm_term": "b",
      "keyword": "poupatempo",
      "device": "m",
      "network": "g",
      "matchtype": "b",
      "gad_source": "1",
      "gad_campaignid": "23455227240",
      "gbraid": "0AAAABCYEGMGzBNrjYICxQ9sUoQhJDcI5n",
      "gclid": "CjwKCAiA4KfLBhB0EiwAUY7GAaMKjCUR8dR-uw5YJKh387ztON6I_BerLAxxCCuYlhPU8ddTndxluxoC8M0QAvD_BwE"
    }
  }'
```

---

### Consultar Pagamento

#### GET `/api/v1/payments/:id`

Consulta um pagamento pelo ID interno.

**Response (200 OK):**

```json
{
  "id": "20a74532-7876-4e3d-9806-5e60130d081a",
  "transaction_id": "d22ed312-0959-45df-b09a-c4c9780b62a2",
  "status": "pending",
  "amount": 10000,
  "payment_method": "pix",
  "platform": "BluPay",
  "pix_code": "00020101021226850014br.gov.bcb.pix...",
  "customer": {
    "id": "615d933c-7a73-44a8-af7e-1059013528fa",
    "name": "Isabel Antônio",
    "email": "isabel_antonio@terra.com.br",
    "phone": "14904767111",
    "document": "45532372101"
  },
  "created_at": "2026-01-17T06:11:53Z",
  "updated_at": "2026-01-17T06:11:53Z"
}
```

**cURL:**

```bash
curl https://server-apis-go-production.up.railway.app/api/v1/payments/20a74532-7876-4e3d-9806-5e60130d081a
```

#### GET `/api/v1/payments/transaction/:transaction_id`

Consulta um pagamento pelo Transaction ID (ID do gateway).

**Response (200 OK):**

```json
{
  "id": "20a74532-7876-4e3d-9806-5e60130d081a",
  "transaction_id": "d22ed312-0959-45df-b09a-c4c9780b62a2",
  "status": "approved",
  "amount": 10000,
  "payment_method": "pix",
  "platform": "BluPay",
  "approved_at": "2026-01-17T06:15:30Z"
}
```

**cURL:**

```bash
curl https://server-apis-go-production.up.railway.app/api/v1/payments/transaction/d22ed312-0959-45df-b09a-c4c9780b62a2
```

---

## Webhooks

### Receber Webhooks

#### POST `/api/v1/webhooks/payment`

Endpoint genérico para receber webhooks de pagamento (suporta todos os gateways).

#### POST `/api/v1/webhooks/blupay`

Endpoint específico para webhooks BluPay.

#### POST `/api/v1/webhooks/quantumpay`

Endpoint específico para webhooks QuantumPay.

**Webhook BluPay (transaction.paid):**

```json
{
  "id": "evt_1765495168629_zhv3arq75",
  "type": "transaction",
  "event": "transaction.paid",
  "objectId": "d22ed312-0959-45df-b09a-c4c9780b62a2",
  "data": {
    "id": "d22ed312-0959-45df-b09a-c4c9780b62a2",
    "status": "paid",
    "amount": 10000,
    "paymentMethod": "PIX",
    "customer": {
      "name": "Isabel Antônio",
      "email": "isabel_antonio@terra.com.br",
      "phone": "14904767111",
      "document": "45532372101"
    },
    "pix": {
      "qrcode": "00020101021226850014br.gov.bcb.pix...",
      "end2EndId": "E18236120202512020455s14af098224"
    },
    "paidAt": "2026-01-17T06:15:30.942Z"
  }
}
```

**Response (200 OK):**

```json
{
  "success": true
}
```

**Headers:**
- `Content-Type: application/json`
- `X-Webhook-Signature: <hmac-sha256>` (quando webhookSecret é fornecido)

**Eventos suportados:**
- `transaction.paid` - Pagamento aprovado
- `transaction.refunded` - Pagamento estornado
- `transaction.cancelled` - Pagamento cancelado

---

### Webhooks Externos

A API pode enviar webhooks para **sua URL** quando um pagamento for aprovado. Basta fornecer o campo `webhook_url` ao criar o pagamento.

#### Como funcionar

1. Ao criar um pagamento (BluPay ou QuantumPay), inclua o campo `webhook_url`:

```json
{
  "amount": 10000,
  "webhook_url": "https://meudominio.com/webhook.php",
  "utm_params": { ... }
}
```

2. Quando o pagamento for **aprovado**, a API enviará automaticamente um webhook para sua URL.

#### Payload do Webhook Externo

```json
{
  "event": "payment.approved",
  "transaction_id": "d22ed312-0959-45df-b09a-c4c9780b62a2",
  "order_id": "20a74532-7876-4e3d-9806-5e60130d081a",
  "status": "approved",
  "amount": 10000,
  "payment_method": "pix",
  "platform": "BluPay",
  "approved_at": "2026-01-17T06:15:30Z",
  "created_at": "2026-01-17T06:11:53Z",
  "customer": {
    "id": "615d933c-7a73-44a8-af7e-1059013528fa",
    "name": "Isabel Antônio",
    "email": "isabel_antonio@terra.com.br",
    "phone": "14904767111",
    "document": "45532372101"
  },
  "tracking_params": {
    "utm_source": "google",
    "utm_campaign": "23455227240",
    "utm_medium": "193094442473",
    "utm_content": "792714546957",
    "utm_term": "b",
    "gclid": "CjwKCAiA4KfLBhB0EiwAUY7GAaMKjCUR8dR...",
    "fbclid": "",
    "ttclid": "",
    "sck": "",
    "xcod": ""
  }
}
```

#### Headers Enviados

- `Content-Type: application/json`
- `User-Agent: Server-APIs-Webhook/1.0`
- `X-Webhook-Event: payment.approved`
- `X-Transaction-ID: d22ed312-0959-45df-b09a-c4c9780b62a2`

#### Retry Policy

A API tenta enviar o webhook até **3 vezes** com intervalo de **60 segundos**:
- Tentativa 1: Imediato
- Tentativa 2: Aguarda 60 segundos
- Tentativa 3: Aguarda 60 segundos

Seu endpoint deve responder com **HTTP 2xx** (200-299) para confirmar recebimento.

#### Exemplo PHP

```php
<?php
// webhook.php

// Lê o payload
$payload = file_get_contents('php://input');
$data = json_decode($payload, true);

// Valida evento
if ($data['event'] === 'payment.approved') {
    $transaction_id = $data['transaction_id'];
    $amount = $data['amount'];
    $customer_email = $data['customer']['email'];
    $utm_source = $data['tracking_params']['utm_source'];

    // Processa o pagamento aprovado
    // ... sua lógica aqui ...

    // Retorna sucesso
    http_response_code(200);
    echo json_encode(['success' => true]);
} else {
    http_response_code(400);
    echo json_encode(['error' => 'Invalid event']);
}
?>
```

#### Exemplo Node.js

```javascript
app.post('/webhook', express.json(), (req, res) => {
  const { event, transaction_id, amount, customer, tracking_params } = req.body;

  if (event === 'payment.approved') {
    console.log(`Pagamento aprovado: ${transaction_id} - R$ ${amount/100}`);
    console.log(`Cliente: ${customer.name} (${customer.email})`);
    console.log(`UTM Source: ${tracking_params.utm_source}`);

    // Processa o pagamento aprovado
    // ... sua lógica aqui ...

    res.json({ success: true });
  } else {
    res.status(400).json({ error: 'Invalid event' });
  }
});
```

#### Testando Webhooks

Para testar localmente, use ferramentas como:
- **ngrok**: `ngrok http 80`
- **webhook.site**: Recebe e exibe webhooks
- **RequestBin**: Inspeciona payloads HTTP

---

## Consulta CPF

### GET `/api/cpf/:cpf`

Consulta dados completos de um CPF.

**Parâmetros:**
- `cpf`: CPF com ou sem formatação

**Response (200 OK):**

```json
{
  "cpf": "123.456.789-10",
  "nome": "JOÃO SILVA",
  "nome_mae": "MARIA SILVA",
  "data_nascimento": "01/01/1990",
  "situacao": "REGULAR",
  "situacao_receita": "REGULAR",
  "ano_obito": "",
  "genero": "MASCULINO",
  "uf": "SP",
  "municipio": "SÃO PAULO",
  "logradouro": "RUA EXEMPLO",
  "numero": "123",
  "complemento": "APTO 45",
  "bairro": "CENTRO",
  "cep": "01234-567",
  "telefone": "(11) 99999-9999",
  "email": "joao@email.com"
}
```

**cURL:**

```bash
curl https://server-apis-go-production.up.railway.app/api/cpf/12345678910
```

### GET `/api/cpf?cpf=12345678910`

Mesma funcionalidade, porém usando query string.

**cURL:**

```bash
curl "https://server-apis-go-production.up.railway.app/api/cpf?cpf=12345678910"
```

---

## Códigos de Status

### Status de Pagamento

- `pending` - Aguardando pagamento
- `waiting_payment` - Aguardando pagamento
- `approved` - Pagamento aprovado
- `paid` - Pagamento confirmado
- `refunded` - Pagamento estornado
- `cancelled` - Pagamento cancelado

### HTTP Status Codes

- `200 OK` - Requisição processada com sucesso
- `201 Created` - Recurso criado com sucesso
- `400 Bad Request` - Dados inválidos na requisição
- `404 Not Found` - Recurso não encontrado
- `500 Internal Server Error` - Erro interno do servidor

---

## Erros Comuns

### 400 Bad Request

```json
{
  "success": false,
  "message": "Dados inválidos: amount must be >= 100"
}
```

**Causas comuns:**
- Valor menor que R$ 1,00 (100 centavos)
- JSON malformado
- Campos obrigatórios ausentes

### 404 Not Found

```json
{
  "error": "Pagamento não encontrado"
}
```

**Causas comuns:**
- ID ou Transaction ID inválido
- Pagamento não existe no banco de dados

### 500 Internal Server Error

```json
{
  "success": false,
  "message": "Erro ao criar pagamento: erro detalhado..."
}
```

**Causas comuns:**
- Erro na comunicação com gateway de pagamento
- Erro no banco de dados
- Erro na integração com Utmify

---

## Funcionalidades Adicionais

### Geração Automática de Dados

Quando os campos `name`, `email`, `document` ou `phone` não são fornecidos, a API gera automaticamente dados válidos usando a API 5devs:

**Request:**
```json
{
  "amount": 10000
}
```

**Response:**
```json
{
  "success": true,
  "token": "d22ed312-0959-45df-b09a-c4c9780b62a2",
  "nome": "Isabel Antônio",
  "cpf": "45532372101",
  "pixCode": "..."
}
```

**Logs do servidor:**
```
🔄 Dados incompletos detectados, gerando automaticamente via 5devs...
✅ Nome gerado: Isabel Antônio
✅ Email gerado: isabel_antonio@terra.com.br
✅ CPF gerado: 45532372101
✅ Telefone gerado: 14904767111
```

### Integração Utmify

Todos os pagamentos são automaticamente enviados para o Utmify com:
- Status inicial: `waiting_payment`
- Status final: `approved` (quando webhook confirma pagamento)
- UTM parameters completos
- Dados de comissão e taxas

### RabbitMQ

Eventos publicados automaticamente:
- `payment.created` - Quando pagamento é criado
- `payment.approved` - Quando pagamento é aprovado
- `utmify.pending` - Para processar envio assíncrono ao Utmify
- `utmify.approved` - Para atualizar status no Utmify

---

## Ambientes

### Produção
- URL: `https://server-apis-go-production.up.railway.app`
- Database: PostgreSQL (Railway)
- Redis: Railway Redis
- RabbitMQ: CloudAMQP

### Desenvolvimento
- URL: `http://localhost:8080`
- Database: PostgreSQL local
- Redis: localhost:6379
- RabbitMQ: localhost:5672

---

## Suporte

Para dúvidas ou problemas:
- GitHub: [caval31r0/server-apis-go](https://github.com/caval31r0/server-apis-go)
- Documentação BluPay: [docs.blupayip.io](https://docs.blupayip.io)
- Documentação QuantumPay: API privada

---

**Última atualização:** 17/01/2026
**Versão:** 1.0.0
