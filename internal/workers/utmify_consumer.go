package workers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/victtorkaiser/server-apis/internal/config"
	"github.com/victtorkaiser/server-apis/internal/models"
	"github.com/victtorkaiser/server-apis/internal/queue"
	"gorm.io/gorm"
)

type UtmifyConsumer struct {
	db       *gorm.DB
	rabbitMQ *queue.RabbitMQ
	cfg      *config.Config
}

func NewUtmifyConsumer(db *gorm.DB, rabbitMQ *queue.RabbitMQ, cfg *config.Config) *UtmifyConsumer {
	return &UtmifyConsumer{
		db:       db,
		rabbitMQ: rabbitMQ,
		cfg:      cfg,
	}
}

func (c *UtmifyConsumer) Start() error {
	log.Println("🚀 Iniciando UtmifyConsumer - processando filas utmify.pending e utmify.approved")

	// Consumer para utmify.pending
	if err := c.rabbitMQ.Consume("utmify.pending", c.handlePendingOrder); err != nil {
		return err
	}

	// Consumer para utmify.approved
	if err := c.rabbitMQ.Consume("utmify.approved", c.handleApprovedOrder); err != nil {
		return err
	}

	log.Println("✅ UtmifyConsumer iniciado com sucesso")
	return nil
}

func (c *UtmifyConsumer) handlePendingOrder(data []byte) error {
	var message map[string]interface{}
	if err := json.Unmarshal(data, &message); err != nil {
		log.Printf("❌ Erro ao decodificar mensagem utmify.pending: %v", err)
		return err
	}

	orderIDStr, ok := message["order_id"].(string)
	if !ok {
		log.Printf("❌ order_id inválido na mensagem: %v", message)
		return nil // Não reprocessa
	}

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		log.Printf("❌ Erro ao parsear order_id: %v", err)
		return nil
	}

	// Busca order completa no banco
	var order models.Order
	if err := c.db.Preload("Customer").Preload("TrackingParameter").Preload("Products").First(&order, "id = ?", orderID).Error; err != nil {
		log.Printf("❌ Erro ao buscar order %s: %v", orderID, err)
		return err
	}

	// Envia para Utmify
	log.Printf("📤 Enviando order %s para Utmify (pending)", order.TransactionID)
	return c.sendToUtmify(&order, "waiting_payment")
}

func (c *UtmifyConsumer) handleApprovedOrder(data []byte) error {
	var message map[string]interface{}
	if err := json.Unmarshal(data, &message); err != nil {
		log.Printf("❌ Erro ao decodificar mensagem utmify.approved: %v", err)
		return err
	}

	orderIDStr, ok := message["order_id"].(string)
	if !ok {
		log.Printf("❌ order_id inválido na mensagem: %v", message)
		return nil
	}

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		log.Printf("❌ Erro ao parsear order_id: %v", err)
		return nil
	}

	// Busca order completa no banco
	var order models.Order
	if err := c.db.Preload("Customer").Preload("TrackingParameter").Preload("Products").First(&order, "id = ?", orderID).Error; err != nil {
		log.Printf("❌ Erro ao buscar order %s: %v", orderID, err)
		return err
	}

	// Envia para Utmify
	log.Printf("📤 Enviando order %s para Utmify (approved)", order.TransactionID)
	return c.sendToUtmify(&order, "paid")
}

func (c *UtmifyConsumer) sendToUtmify(order *models.Order, status string) error {
	if c.cfg.UtmifyAPIURL == "" || c.cfg.UtmifyToken == "" {
		log.Println("⚠️ Utmify não configurado, pulando envio")
		return nil
	}

	// Prepara tracking parameters (todos os campos obrigatórios no Utmify como string ou null)
	trackingParams := map[string]interface{}{
		"utm_source":   getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.UtmSource }),
		"utm_medium":   getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.UtmMedium }),
		"utm_campaign": getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.UtmCampaign }),
		"utm_content":  getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.UtmContent }),
		"utm_term":     getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.UtmTerm }),
		"src":          getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.Src }),
		"sck":          getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.Sck }),
		"xcod":         getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.Xcod }),
		"fbclid":       getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.Fbclid }),
		"gclid":        getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.Gclid }),
		"ttclid":       getStringOrNil(order.TrackingParameter, func(tp *models.TrackingParameter) string { return tp.Ttclid }),
	}

	// Prepara produtos
	products := []map[string]interface{}{}
	if len(order.Products) > 0 {
		for _, p := range order.Products {
			products = append(products, map[string]interface{}{
				"id":           p.Code,
				"name":         p.Name,
				"planId":       p.PlanID,
				"planName":     p.PlanName,
				"quantity":     p.Quantity,
				"priceInCents": p.Price,
			})
		}
	} else {
		// Produto padrão
		productName := "Produto"
		if order.Platform == "QuantumPay" && c.cfg.QuantumPayProductName != "" {
			productName = c.cfg.QuantumPayProductName
		}
		products = append(products, map[string]interface{}{
			"id":           "PROD_" + order.TransactionID,
			"name":         productName,
			"planId":       nil,
			"planName":     nil,
			"quantity":     1,
			"priceInCents": order.Amount,
		})
	}

	// Datas
	var approvedDate interface{} = nil
	if order.ApprovedAt != nil {
		approvedDate = order.ApprovedAt.Format("2006-01-02 15:04:05")
	}

	var refundedAt interface{} = nil
	if order.RefundedAt != nil {
		refundedAt = order.RefundedAt.Format("2006-01-02 15:04:05")
	}

	// Monta payload
	payload := map[string]interface{}{
		"orderId":       order.TransactionID,
		"platform":      order.Platform,
		"paymentMethod": "pix",
		"status":        status,
		"createdAt":     order.CreatedAt.Format("2006-01-02 15:04:05"),
		"approvedDate":  approvedDate,
		"refundedAt":    refundedAt,
		"customer": map[string]interface{}{
			"name":     order.Customer.Name,
			"email":    order.Customer.Email,
			"phone":    order.Customer.Phone,
			"document": order.Customer.Document,
			"country":  order.Customer.Country,
			"ip":       order.Customer.IP,
		},
		"products":           products,
		"trackingParameters": trackingParams,
		"commission": map[string]interface{}{
			"totalPriceInCents":     order.Amount,
			"gatewayFeeInCents":     0,
			"userCommissionInCents": order.Amount,
		},
		"isTest": false,
	}

	// Envia para Utmify
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", c.cfg.UtmifyAPIURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("❌ Erro ao criar requisição Utmify: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-token", c.cfg.UtmifyToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Erro ao enviar para Utmify: %v", err)
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📡 [Utmify %s] HTTP %d: %s", status, resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Resposta não-200 do Utmify: %s", string(respBody))
		return nil // Não reprocessa erros de API
	}

	log.Printf("✅ Order %s enviada com sucesso para Utmify (%s)", order.TransactionID, status)
	return nil
}

// Helper para extrair string de TrackingParameter ou retornar nil
func getStringOrNil(tp *models.TrackingParameter, getter func(*models.TrackingParameter) string) interface{} {
	if tp == nil {
		return nil
	}
	val := getter(tp)
	if val == "" {
		return nil
	}
	return val
}
