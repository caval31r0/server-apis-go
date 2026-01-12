package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending        OrderStatus = "pending"
	OrderStatusWaitingPayment OrderStatus = "waiting_payment"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusApproved       OrderStatus = "approved"
	OrderStatusRefunded       OrderStatus = "refunded"
	OrderStatusCancelled      OrderStatus = "cancelled"
)

type Order struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	TransactionID string     `gorm:"uniqueIndex;not null" json:"transaction_id"`
	Status        OrderStatus `gorm:"type:varchar(50);not null" json:"status"`
	Amount        int        `gorm:"not null" json:"amount"` // em centavos
	PaymentMethod string     `gorm:"type:varchar(50)" json:"payment_method"`
	Platform      string     `gorm:"type:varchar(100)" json:"platform"`
	PixCode       string     `gorm:"type:text" json:"pix_code,omitempty"`

	CustomerID uuid.UUID `gorm:"type:uuid" json:"customer_id"`
	Customer   Customer  `gorm:"foreignKey:CustomerID" json:"customer"`

	Products []Product `gorm:"many2many:order_products;" json:"products"`

	TrackingParameterID *uuid.UUID         `gorm:"type:uuid" json:"tracking_parameter_id,omitempty"`
	TrackingParameter   *TrackingParameter `gorm:"foreignKey:TrackingParameterID" json:"tracking_parameters,omitempty"`

	ApprovedAt *time.Time `json:"approved_at,omitempty"`
	RefundedAt *time.Time `json:"refunded_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

type Customer struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name     string    `gorm:"type:varchar(255);not null" json:"name"`
	Email    string    `gorm:"type:varchar(255);not null" json:"email"`
	Phone    string    `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Document string    `gorm:"type:varchar(20);not null" json:"document"`
	Country  string    `gorm:"type:varchar(2);default:'BR'" json:"country"`
	IP       string    `gorm:"type:varchar(45)" json:"ip,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type Product struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Code      string    `gorm:"type:varchar(100);not null" json:"code"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	PlanID    string    `gorm:"type:varchar(100)" json:"plan_id,omitempty"`
	PlanName  string    `gorm:"type:varchar(255)" json:"plan_name,omitempty"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	Price     int       `gorm:"not null" json:"price"` // em centavos

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type TrackingParameter struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Src         string    `gorm:"type:varchar(255)" json:"src,omitempty"`
	Sck         string    `gorm:"type:varchar(255)" json:"sck,omitempty"`
	UtmSource   string    `gorm:"type:varchar(255)" json:"utm_source,omitempty"`
	UtmCampaign string    `gorm:"type:varchar(255)" json:"utm_campaign,omitempty"`
	UtmMedium   string    `gorm:"type:varchar(255)" json:"utm_medium,omitempty"`
	UtmContent  string    `gorm:"type:varchar(255)" json:"utm_content,omitempty"`
	UtmTerm     string    `gorm:"type:varchar(255)" json:"utm_term,omitempty"`
	Xcod        string    `gorm:"type:varchar(255)" json:"xcod,omitempty"`
	Fbclid      string    `gorm:"type:varchar(255)" json:"fbclid,omitempty"`
	Gclid       string    `gorm:"type:varchar(255)" json:"gclid,omitempty"`
	Ttclid      string    `gorm:"type:varchar(255)" json:"ttclid,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *TrackingParameter) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
