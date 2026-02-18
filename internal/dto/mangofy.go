package dto

// MangoFy Request DTOs
type MangoFyRequest struct {
	Amount     int                    `json:"amount" binding:"required,min=1"`
	Name       string                 `json:"name"`
	Email      string                 `json:"email"`
	Document   string                 `json:"document"`
	Phone      string                 `json:"phone"`
	IP         string                 `json:"ip"`
	WebhookURL string                 `json:"webhook_url"`
	UTMParams  map[string]interface{} `json:"utm_params"`
}

// MangoFy API Response DTO
type MangoFyAPIResponse struct {
	PaymentCode string `json:"payment_code"`
	PixCode     string `json:"pix_code"`
	Pix         struct {
		PixQRCodeText string `json:"pix_qrcode_text"`
		PixLink       string `json:"pix_link"`
	} `json:"pix"`
}

// MangoFy Response DTO
type MangoFyResponse struct {
	Success   bool   `json:"success"`
	Token     string `json:"token"`
	PixCode   string `json:"pix_code"`
	QRCodeURL string `json:"qr_code_url"`
	Amount    int    `json:"amount"`
	Nome      string `json:"nome"`
	CPF       string `json:"cpf"`
}
