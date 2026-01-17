package dto

// BluPay Request DTOs
type BluPayRequest struct {
	Amount       int                    `json:"amount" binding:"required,min=100"`
	Name         string                 `json:"name"`
	Email        string                 `json:"email"`
	Document     string                 `json:"document"`
	Phone        string                 `json:"phone"`
	ExternalRef  string                 `json:"externalRef"`
	UTMParams    map[string]interface{} `json:"utm_params"`
}

type BluPayAPIRequest struct {
	Amount        int                 `json:"amount"`
	PaymentMethod string              `json:"paymentMethod"`
	ExternalRef   string              `json:"externalRef"`
	Customer      BluPayCustomer      `json:"customer"`
	Items         []BluPayItem        `json:"items"`
	PostbackUrl   string              `json:"postbackUrl"`
	WebhookSecret string              `json:"webhookSecret"`
	Metadata      map[string]string   `json:"metadata"`
}

type BluPayCustomer struct {
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	Document BluPayDocument  `json:"document"`
}

type BluPayDocument struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type BluPayItem struct {
	Title     string `json:"title"`
	UnitPrice int    `json:"unitPrice"`
	Quantity  int    `json:"quantity"`
	Tangible  bool   `json:"tangible"`
}

// BluPay Response DTOs
type BluPayResponse struct {
	Success   bool   `json:"success"`
	Token     string `json:"token"`
	PixCode   string `json:"pixCode"`
	QRCodeURL string `json:"qrCodeUrl"`
	Amount    int    `json:"amount"`
	Nome      string `json:"nome"`
	CPF       string `json:"cpf"`
	ExpiraEm  string `json:"expiraEm"`
}

type BluPayAPIResponse struct {
	ID              string           `json:"id"`
	Amount          int              `json:"amount"`
	RefundedAmount  int              `json:"refundedAmount"`
	CompanyID       string           `json:"companyId"`
	PaymentMethod   string           `json:"paymentMethod"`
	Status          string           `json:"status"`
	ExternalRef     string           `json:"externalRef"`
	PostbackUrl     string           `json:"postbackUrl"`
	Metadata        interface{}      `json:"metadata"`
	Traceable       bool             `json:"traceable"`
	SecureID        string           `json:"secureId"`
	SecureUrl       string           `json:"secureUrl"`
	CreatedAt       string           `json:"createdAt"`
	UpdatedAt       string           `json:"updatedAt"`
	PaidAt          interface{}      `json:"paidAt"`
	IP              string           `json:"ip"`
	Customer        BluPayCustomerResp `json:"customer"`
	Pix             BluPayPixResp    `json:"pix"`
	Card            interface{}      `json:"card"`
	Boleto          interface{}      `json:"boleto"`
	Shipping        interface{}      `json:"shipping"`
	RefusedReason   interface{}      `json:"refusedReason"`
	Items           []BluPayItemResp `json:"items"`
	Splits          []interface{}    `json:"splits"`
	Refunds         []interface{}    `json:"refunds"`
	Delivery        interface{}      `json:"delivery"`
	Fee             BluPayFee        `json:"fee"`
}

type BluPayCustomerResp struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Phone     string          `json:"phone"`
	CreatedAt string          `json:"createdAt"`
	Document  BluPayDocument  `json:"document"`
}

type BluPayPixResp struct {
	QRCode    string `json:"qrcode"`
	ExpiresAt string `json:"expiresAt"`
}

type BluPayItemResp struct {
	ExternalRef interface{} `json:"externalRef"`
	Title       string      `json:"title"`
	UnitPrice   int         `json:"unitPrice"`
	Quantity    int         `json:"quantity"`
	Tangible    bool        `json:"tangible"`
}

type BluPayFee struct {
	FixedAmount       int     `json:"fixedAmount"`
	SpreadPercentage  float64 `json:"spreadPercentage"`
	EstimatedFee      int     `json:"estimatedFee"`
	NetAmount         int     `json:"netAmount"`
	PixInFeeType      string  `json:"pixInFeeType"`
}

// BluPay Webhook DTOs
type BluPayWebhook struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Event    string                 `json:"event"`
	ObjectID string                 `json:"objectId"`
	Data     BluPayWebhookData      `json:"data"`
}

type BluPayWebhookData struct {
	ID             string                 `json:"id"`
	Status         string                 `json:"status"`
	Amount         int                    `json:"amount"`
	RefundedAmount int                    `json:"refundedAmount"`
	Installments   int                    `json:"installments"`
	PaymentMethod  string                 `json:"paymentMethod"`
	CompanyID      string                 `json:"companyId"`
	ExternalRef    string                 `json:"externalRef"`
	Customer       BluPayWebhookCustomer  `json:"customer"`
	Pix            BluPayWebhookPix       `json:"pix"`
	PaidAt         string                 `json:"paidAt"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
	PostbackUrl    string                 `json:"postbackUrl"`
}

type BluPayWebhookCustomer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Document  string `json:"document"`
	CreatedAt string `json:"createdAt"`
}

type BluPayWebhookPix struct {
	QRCode    string                  `json:"qrcode"`
	End2EndID string                  `json:"end2EndId"`
	Payer     BluPayWebhookPixPayer   `json:"payer"`
}

type BluPayWebhookPixPayer struct {
	Name         string                       `json:"name"`
	Document     string                       `json:"document"`
	DocumentType string                       `json:"documentType"`
	BankAccount  BluPayWebhookPixBankAccount  `json:"bankAccount"`
}

type BluPayWebhookPixBankAccount struct {
	ISPB    string `json:"ispb"`
	Branch  string `json:"branch"`
	Account string `json:"account"`
}
