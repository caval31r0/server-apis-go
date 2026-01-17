# API BluPay - Integração PIX

## Visão Geral

Integração completa com a API BluPayIP para criação de pagamentos PIX com suporte a:
- ✅ Criação de cobranças PIX
- ✅ Geração automática de QR Code
- ✅ Rastreamento de UTM parameters
- ✅ Webhooks para confirmação de pagamento
- ✅ Integração com Utmify
- ✅ Geração automática de dados de teste (via 5devs)

## Endpoint

### POST `/api/payment/blupay`

Cria uma nova cobrança PIX através da BluPay.

#### Request Body

```json
{
  "amount": 10000,
  "name": "João Silva",
  "email": "joao@email.com",
  "document": "12345678910",
  "phone": "11999999999",
  "externalRef": "ORD-123",
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

#### Campos Obrigatórios
- `amount`: Valor em centavos (mínimo: 100 = R$ 1,00)

#### Campos Opcionais (auto-gerados se não fornecidos)
- `name`: Nome do cliente
- `email`: Email do cliente
- `document`: CPF do cliente (apenas números)
- `phone`: Telefone com DDD (apenas números)
- `externalRef`: Referência externa (gerado automaticamente se não fornecido)
- `utm_params`: Parâmetros de rastreamento UTM

#### Response Success (200 OK)

```json
{
  "success": true,
  "token": "bc06ccc9-c64f-4dc5-b54e-baabf08fbb1b",
  "pixCode": "00020101021226880014br.gov.bcb.pix2566qrcode.example.com...",
  "qrCodeUrl": "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=...",
  "amount": 10000,
  "nome": "João Silva",
  "cpf": "12345678910",
  "expiraEm": "2 dias"
}
```

#### Response Error (400 Bad Request)

```json
{
  "success": false,
  "message": "Dados inválidos: amount must be >= 100"
}
```

#### Response Error (500 Internal Server Error)

```json
{
  "success": false,
  "message": "Erro ao criar pagamento: erro detalhado..."
}
```

## Configuração

### Variáveis de Ambiente

Adicione as seguintes variáveis ao seu arquivo `.env`:

```bash
# BluPay Configuration
BLUPAY_API_URL=https://api.blupayip.io/api/v1
BLUPAY_SECRET_KEY=live_-8EI6hKJSkaYUyvyBjBlDZdkfee0hY8_
BLUPAY_PUBLIC_KEY=65136884-dd99-4ede-8566-28505082473a
BLUPAY_WEBHOOK_SECRET=secret_900de97d1cf10dda70c803fede642899
BLUPAY_WEBHOOK_URL=https://seu-dominio.com/api/v1/webhooks/blupay
BLUPAY_PRODUCT_NAME=Produto Digital
```

**IMPORTANTE**: Configure a URL do webhook no painel da BluPay:
- URL: `https://seu-dominio.com/api/v1/webhooks/blupay`
- Eventos: `transaction.paid`, `transaction.refunded`, `transaction.cancelled`

### Credenciais BluPay

- **Chave Pública (Company ID)**: `65136884-dd99-4ede-8566-28505082473a`
- **Chave Secreta**: `live_-8EI6hKJSkaYUyvyBjBlDZdkfee0hY8_`
- **Webhook Secret**: `secret_900de97d1cf10dda70c803fede642899`

## Webhooks

A BluPay envia webhooks para confirmar pagamentos. Configure a URL de webhook nas variáveis de ambiente.

### Evento: `transaction.paid`

Webhook enviado quando um pagamento PIX é confirmado.

#### Headers

- `Content-Type`: `application/json`
- `X-Webhook-Signature`: HMAC-SHA256 do body (quando `webhookSecret` é fornecido)

#### Payload

```json
{
  "id": "evt_1765495168629_zhv3arq75",
  "type": "transaction",
  "event": "transaction.paid",
  "objectId": "0b922550-ebf7-433b-a5d9-ee56a3c38285",
  "data": {
    "id": "0b922550-ebf7-433b-a5d9-ee56a3c38285",
    "status": "paid",
    "amount": 10000,
    "refundedAmount": 0,
    "installments": 1,
    "paymentMethod": "PIX",
    "companyId": "4d1a3c25-2cfc-4f72-b814-23d9fd168c8e",
    "externalRef": "ORD-123",
    "customer": {
      "id": "e00d270d-84de-4b94-8d1b-3bb26921e04f",
      "name": "João Silva",
      "email": "joao@email.com",
      "phone": "11999999999",
      "document": "12345678910",
      "createdAt": "2025-12-02T19:12:51.628Z"
    },
    "pix": {
      "qrcode": "00020101021226880014br.gov.bcb.pix...",
      "end2EndId": "E18236120202512020455s14af098224",
      "payer": {
        "name": "Carlos Eduardo Santos",
        "document": "98765432100",
        "documentType": "cpf",
        "bankAccount": {
          "ispb": "18236120",
          "branch": "1",
          "account": "123456789"
        }
      }
    },
    "paidAt": "2025-12-02T04:55:26.942Z",
    "createdAt": "2025-12-11T23:16:24.094Z",
    "updatedAt": "2025-12-11T23:19:28.603Z",
    "postbackUrl": "https://seu-dominio.com/api/webhooks/blupay"
  }
}
```

## Exemplos de Uso

### cURL - Pagamento Completo

```bash
curl -X POST http://localhost:8080/api/payment/blupay \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 10000,
    "name": "João Silva",
    "email": "joao@email.com",
    "document": "12345678910",
    "phone": "11999999999",
    "externalRef": "ORD-123",
    "utm_params": {
      "utm_source": "google",
      "utm_medium": "cpc",
      "utm_campaign": "black-friday"
    }
  }'
```

### cURL - Apenas Amount (dados auto-gerados)

```bash
curl -X POST http://localhost:8080/api/payment/blupay \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 5000
  }'
```

### JavaScript/Fetch

```javascript
const response = await fetch('http://localhost:8080/api/payment/blupay', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    amount: 10000,
    name: 'João Silva',
    email: 'joao@email.com',
    document: '12345678910',
    phone: '11999999999',
    utm_params: {
      utm_source: 'google',
      utm_medium: 'cpc',
      utm_campaign: 'black-friday'
    }
  })
});

const data = await response.json();
console.log(data);
```

### Python/Requests

```python
import requests

payload = {
    "amount": 10000,
    "name": "João Silva",
    "email": "joao@email.com",
    "document": "12345678910",
    "phone": "11999999999",
    "utm_params": {
        "utm_source": "google",
        "utm_medium": "cpc",
        "utm_campaign": "black-friday"
    }
}

response = requests.post(
    'http://localhost:8080/api/payment/blupay',
    json=payload
)

print(response.json())
```

## Recursos

### Geração Automática de Dados

Se os campos `name`, `email`, `document` ou `phone` não forem fornecidos, a API gera automaticamente usando a API 5devs:

```bash
curl -X POST http://localhost:8080/api/payment/blupay \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000}'
```

Logs gerados:
```
🔄 Dados incompletos detectados, gerando automaticamente via 5devs...
✅ Nome gerado: Carlos Eduardo Silva
✅ Email gerado: carlos.silva@email.com
✅ CPF gerado: 12345678910
✅ Telefone gerado: 11999887766
```

### Integração com Utmify

Todos os pagamentos são automaticamente enviados para o Utmify (se configurado) com status `waiting_payment`.

Quando o pagamento é confirmado via webhook, o status é atualizado para `approved`.

### RabbitMQ

Eventos publicados:
- `payment.created` - Quando um pagamento é criado
- `utmify.pending` - Para processar envio assíncrono ao Utmify

## Fluxo de Pagamento

1. Cliente faz POST para `/api/payment/blupay`
2. API valida dados e gera dados faltantes (se necessário)
3. API cria customer no banco de dados
4. API chama BluPay API para criar transação PIX
5. API salva order no banco de dados
6. API envia evento para RabbitMQ
7. API envia dados para Utmify (assíncrono)
8. API retorna PIX code e QR code para o cliente
9. Cliente escaneia QR code e paga
10. BluPay envia webhook confirmando pagamento
11. API atualiza status e notifica Utmify

## Status da Transação

- `pending` - Aguardando pagamento
- `paid` - Pago
- `cancelled` - Cancelado
- `refunded` - Estornado

## Documentação Oficial

- [BluPay API Docs](https://docs.blupayip.io/)

## Suporte

Para dúvidas sobre a integração BluPay, consulte a documentação oficial ou entre em contato com o suporte BluPay.
