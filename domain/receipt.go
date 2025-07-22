package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReceiptItem struct {
	Brand        string  `json:"brand" validate:"required"`
	MedicineName string  `json:"medicine_name" validate:"required"`
	PricePerUnit float64 `json:"price_per_unit" validate:"required"`
	Quantity     int     `json:"quantity" validate:"required"`
	Subtotal     float64 `json:"subtotal" validate:"required"`
}

type ReceiptContent struct {
	Items      []ReceiptItem `json:"items" validate:"required"`
	PharmacyID uuid.UUID     `json:"pharmacy_id" validate:"required"`
	SaleDate   time.Time     `json:"sale_date" validate:"required"`
	TotalPrice float64       `json:"total_price" validate:"required"`
}

type Receipt struct {
	ID        uuid.UUID      `json:"id" validate:"required"`
	SaleID    uuid.UUID      `json:"sale_id" validate:"required"`
	Content   ReceiptContent `json:"content" validate:"required"`
	CreatedAt time.Time      `json:"created_at" validate:"required"`
}

type ReceiptResponse struct {
	ID         uuid.UUID     `json:"id"`
	SaleID     uuid.UUID     `json:"sale_id"`
	Items      []ReceiptItem `json:"items"`
	PharmacyID uuid.UUID     `json:"pharmacy_id"`
	SaleDate   time.Time     `json:"sale_date"`
	TotalPrice float64       `json:"total_price"`
	CreatedAt  time.Time     `json:"created_at"`
}
