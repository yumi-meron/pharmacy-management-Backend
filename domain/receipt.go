package domain

import (
	"time"

	"github.com/google/uuid"
)

type Receipt struct {
	ID            uuid.UUID     `json:"id"`
	ReceiptNumber string        `json:"receipt_number"`
	PharmacyId    uuid.UUID     `json:"pharmacy_id"`
	PharmacistId  uuid.UUID     `json:"pharmacist_id"`
	Datetime      time.Time     `json:"datetime"`
	TotalAmount   float64       `json:"total_amount"`
	Items         []ReceiptItem `json:"items"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type ReceiptItem struct {
	ID                uuid.UUID `json:"id"`
	ReceiptID         uuid.UUID `json:"receipt_id"`
	MedicineVariantID uuid.UUID `json:"medicine_variant_id"`
	Quantity          int       `json:"quantity"`
	PricePerUnit      float64   `json:"price_per_unit"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
