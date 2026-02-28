package dto

// Genesys Request DTOs
type GenesysRequest struct {
	Amount     int                    `json:"amount" binding:"required,min=1"`
	Name       string                 `json:"name"`
	Email      string                 `json:"email"`
	Document   string                 `json:"document"`
	Phone      string                 `json:"phone"`
	IP         string                 `json:"ip"`
	WebhookURL string                 `json:"webhook_url"`
	UTMParams  map[string]interface{} `json:"utm_params"`
}

// Genesys API Request (enviado para api.genesys.finance)
type GenesysAPIRequest struct {
	ExternalID    string                `json:"external_id"`
	TotalAmount   float64               `json:"total_amount"`
	PaymentMethod string                `json:"payment_method"`
	WebhookURL    string                `json:"webhook_url"`
	Items         []GenesysItem         `json:"items"`
	IP            string                `json:"ip,omitempty"`
	Customer      GenesysCustomer       `json:"customer"`
}

type GenesysItem struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	IsPhysical  bool    `json:"is_physical"`
}

type GenesysCustomer struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	DocumentType string `json:"document_type"`
	Document     string `json:"document"`
}

// Genesys API Response
type GenesysAPIResponse struct {
	ID            string         `json:"id"`
	ExternalID    string         `json:"external_id"`
	Status        string         `json:"status"`
	TotalValue    float64        `json:"total_value"`
	PaymentMethod string         `json:"payment_method"`
	Pix           GenesysPixResp `json:"pix"`
	HasError      bool           `json:"hasError"`
	Customer      struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"customer"`
}

type GenesysPixResp struct {
	Payload string `json:"payload"`
}

// Genesys Response (retornado ao nosso cliente)
type GenesysResponse struct {
	Success   bool   `json:"success"`
	Token     string `json:"token"`
	PixCode   string `json:"pix_code"`
	QRCodeURL string `json:"qr_code_url"`
	Amount    int    `json:"amount"`
	Nome      string `json:"nome"`
	CPF       string `json:"cpf"`
}

// Genesys Webhook Payload
type GenesysWebhookPayload struct {
	ID            string  `json:"id"`
	ExternalID    string  `json:"external_id"`
	TotalAmount   float64 `json:"total_amount"`
	Status        string  `json:"status"`
	PaymentMethod string  `json:"payment_method"`
}
