package dto

import "fmt"

type CreatePaymentRequest struct {
	Amount     int                    `json:"amount" binding:"required,min=1"`
	Name       string                 `json:"name"`
	Email      string                 `json:"email"`
	Document   string                 `json:"document"`
	Telephone  string                 `json:"telephone"`
	UTMParams  map[string]interface{} `json:"utm_params"`
}

type CreatePaymentResponse struct {
	Success   bool   `json:"success"`
	Token     string `json:"token"`
	PixCode   string `json:"pix_code,omitempty"`
	QRCodeURL string `json:"qr_code_url,omitempty"`
	Amount    int    `json:"amount"`
}

// WebhookPayload - Suporta múltiplos formatos de webhook
type WebhookPayload struct {
	// Formato QuantumPay
	Type     string              `json:"type"`
	ObjectID interface{}         `json:"objectId"` // Pode ser string ou número
	Data     *WebhookDataPayload `json:"data"`

	// Formato BluPay
	ID    string `json:"id"`    // ID do evento
	Event string `json:"event"` // transaction.paid, transaction.refunded, etc

	// Formato legado MangoFy
	PaymentCode   string `json:"payment_code"`
	PaymentStatus string `json:"payment_status"`
	PaymentID     string `json:"paymentId"`
	Status        string `json:"status"`
}

type WebhookDataPayload struct {
	ID            interface{}            `json:"id"`
	Status        string                 `json:"status"`
	Amount        int                    `json:"amount"`
	PaymentMethod string                 `json:"paymentMethod"`
	PaidAt        *string                `json:"paidAt"`
	CreatedAt     string                 `json:"createdAt"`
	Customer      *WebhookCustomer       `json:"customer"`
	Items         []WebhookItem          `json:"items"`
	Fee           *WebhookFee            `json:"fee"`
	Metadata      string                 `json:"metadata"`
}

type WebhookCustomer struct {
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	Document *WebhookDocument `json:"document"`
}

type WebhookDocument struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type WebhookItem struct {
	Title       string `json:"title"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int    `json:"unitPrice"`
	ExternalRef string `json:"externalRef"`
}

type WebhookFee struct {
	NetAmount    int `json:"netAmount"`
	FixedAmount  int `json:"fixedAmount"`
	EstimatedFee int `json:"estimatedFee"`
}

func (w *WebhookPayload) GetPaymentCode() string {
	// Formato BluPay (usa objectId do webhook)
	if w.Event != "" && w.ObjectID != nil {
		switch v := w.ObjectID.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%.0f", v)
		case int:
			return fmt.Sprintf("%d", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	// Formato QuantumPay
	if w.ObjectID != nil {
		switch v := w.ObjectID.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%.0f", v)
		case int:
			return fmt.Sprintf("%d", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	// Formato BluPay - data.id
	if w.Data != nil && w.Data.ID != nil {
		switch v := w.Data.ID.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%.0f", v)
		case int:
			return fmt.Sprintf("%d", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	// Formato legado
	if w.PaymentCode != "" {
		return w.PaymentCode
	}
	return w.PaymentID
}

func (w *WebhookPayload) GetStatus() string {
	// Formato BluPay (extrai status do event)
	if w.Event != "" {
		switch w.Event {
		case "transaction.paid":
			return "paid"
		case "transaction.refunded":
			return "refunded"
		case "transaction.cancelled":
			return "cancelled"
		}
	}

	// Formato QuantumPay / BluPay data.status
	if w.Data != nil && w.Data.Status != "" {
		return w.Data.Status
	}

	// Formato legado
	if w.PaymentStatus != "" {
		return w.PaymentStatus
	}
	return w.Status
}
