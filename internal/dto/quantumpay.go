package dto

// QuantumPay Request DTOs
type QuantumPayRequest struct {
	Amount     int                    `json:"amount" binding:"required,min=1"`
	Name       string                 `json:"name"`
	Email      string                 `json:"email"`
	Document   string                 `json:"document"`
	Telephone  string                 `json:"telephone"`
	WebhookURL string                 `json:"webhook_url"`
	UTMParams  map[string]interface{} `json:"utm_params"`
}

type QuantumPayAPIRequest struct {
	Amount        int                    `json:"amount"`
	PaymentMethod string                 `json:"paymentMethod"`
	Pix           QuantumPayPixConfig    `json:"pix"`
	Customer      QuantumPayCustomer     `json:"customer"`
	Items         []QuantumPayItem       `json:"items"`
	Metadata      string                 `json:"metadata"`
	IP            string                 `json:"ip"`
}

type QuantumPayPixConfig struct {
	ExpiresInDays int `json:"expiresInDays"`
}

type QuantumPayCustomer struct {
	Name        string                 `json:"name"`
	Email       string                 `json:"email"`
	Phone       string                 `json:"phone"`
	Document    QuantumPayDocument     `json:"document"`
	ExternalRef string                 `json:"externalRef"`
}

type QuantumPayDocument struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type QuantumPayItem struct {
	Title       string `json:"title"`
	UnitPrice   int    `json:"unitPrice"`
	Quantity    int    `json:"quantity"`
	Tangible    bool   `json:"tangible"`
	ExternalRef string `json:"externalRef"`
}

// QuantumPay Response DTOs
type QuantumPayResponse struct {
	Success   bool   `json:"success"`
	Token     string `json:"token"`
	PixCode   string `json:"pixCode"`
	QRCodeURL string `json:"qrCodeUrl"`
	Amount    int    `json:"amount"`
	Nome      string `json:"nome"`
	CPF       string `json:"cpf"`
	ExpiraEm  string `json:"expiraEm"`
	Txid      string `json:"txid,omitempty"`
}

type QuantumPayAPIResponse struct {
	ID  interface{}         `json:"id"` // Pode ser string ou número
	Pix QuantumPayPixResult `json:"pix"`
	Fee struct {
		Amount int `json:"amount"`
	} `json:"fee"`
}

type QuantumPayPixResult struct {
	QRCode     string `json:"qrcode"`
	QRCodeURL  string `json:"qrcodeUrl"`
	ReceiptURL string `json:"receiptUrl"`
	End2EndID  string `json:"end2EndId"`
	Txid       string `json:"txid"`
}
