package dto

import "time"

type UtmifyOrderRequest struct {
	OrderID       string                 `json:"orderId"`
	Platform      string                 `json:"platform"`
	PaymentMethod string                 `json:"paymentMethod"`
	Status        string                 `json:"status"`
	CreatedAt     time.Time              `json:"createdAt"`
	ApprovedDate  *time.Time             `json:"approvedDate,omitempty"`
	PaidAt        *time.Time             `json:"paidAt,omitempty"`
	RefundedAt    *time.Time             `json:"refundedAt,omitempty"`
	Customer      UtmifyCustomer         `json:"customer"`
	Products      []UtmifyProduct        `json:"products,omitempty"`
	Items         []UtmifyItem           `json:"items,omitempty"`
	Amount        int                    `json:"amount,omitempty"`
	Fee           *UtmifyFee             `json:"fee,omitempty"`
	Commission    *UtmifyCommission      `json:"commission,omitempty"`
	TrackingParams map[string]interface{} `json:"trackingParameters"`
	IsTest        bool                   `json:"isTest"`
}

type UtmifyCustomer struct {
	Name     string              `json:"name"`
	Email    string              `json:"email"`
	Phone    string              `json:"phone,omitempty"`
	Document interface{}         `json:"document"` // pode ser string ou objeto
	Country  string              `json:"country"`
	IP       string              `json:"ip,omitempty"`
}

type UtmifyProduct struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PlanID    string `json:"planId,omitempty"`
	PlanName  string `json:"planName,omitempty"`
	Quantity  int    `json:"quantity"`
	Price     int    `json:"priceInCents"`
}

type UtmifyItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Quantity  int    `json:"quantity"`
	UnitPrice int    `json:"unitPrice"`
}

type UtmifyFee struct {
	FixedAmount int `json:"fixedAmount"`
	NetAmount   int `json:"netAmount"`
}

type UtmifyCommission struct {
	TotalPrice        int `json:"totalPriceInCents"`
	GatewayFee        int `json:"gatewayFeeInCents"`
	UserCommission    int `json:"userCommissionInCents"`
}
