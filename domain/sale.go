package domain

import (
	"time"

	"github.com/google/uuid"
)

// Cart represents an item in the user's cart
type Cart struct {
	ID                uuid.UUID `json:"id" validate:"required"`
	UserID            uuid.UUID `json:"user_id" validate:"required"`
	PharmacyID        uuid.UUID `json:"pharmacy_id" validate:"required"`
	MedicineVariantID uuid.UUID `json:"medicine_variant_id" validate:"required"`
	Quantity          int       `json:"quantity" validate:"required,gt=0"`
	CreatedAt         time.Time `json:"created_at" validate:"required"`
	// Temporary fields for response
	MedicineName string  `json:"medicine,omitempty"`
	PricePerUnit float64 `json:"price_per_unit,omitempty"`
	Unit         string  `json:"unit,omitempty"`
	ImageURL     string  `json:"image_url,omitempty"`
}

// CreateCartInput for adding an item to the cart
type CreateCartInput struct {
	MedicineVariantID uuid.UUID `json:"medicine_variant_id" validate:"required"`
	Quantity          int       `json:"quantity" validate:"required,gt=0"`
}

// CartResponse represents the response structure for a cart item
type CartResponse struct {
	ID           uuid.UUID `json:"id"`
	Medicine     string    `json:"medicine"`
	PricePerUnit float64   `json:"price_per_unit"`
	Unit         string    `json:"unit"`
	ImageURL     string    `json:"image_url"`
	Quantity     int       `json:"quantity"`
	CreatedAt    time.Time `json:"created_at"`
}

// SaleItem represents an item in a sale
type SaleItem struct {
	ID                uuid.UUID `json:"id" validate:"required"`
	SaleID            uuid.UUID `json:"sale_id" validate:"required"`
	MedicineVariantID uuid.UUID `json:"medicine_variant_id" validate:"required"`
	Quantity          int       `json:"quantity" validate:"required,gt=0"`
	PricePerUnit      float64   `json:"price_per_unit" validate:"required,gt=0"`
	CreatedAt         time.Time `json:"created_at" validate:"required"`
	// Temporary fields for response
	MedicineName string `json:"medicine,omitempty"`
	Unit         string `json:"unit,omitempty"`
	ImageURL     string `json:"image_url,omitempty"`
}

// Sale represents a completed sale
type Sale struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	PharmacyID uuid.UUID `json:"pharmacy_id" validate:"required"`
	TotalPrice float64   `json:"total_price" validate:"required,gte=0"`
	SaleDate   time.Time `json:"sale_date" validate:"required"`
	CreatedAt  time.Time `json:"created_at" validate:"required"`
	UpdatedAt  time.Time `json:"updated_at" validate:"required"`
}

// SaleResponse represents the response structure for a sale
type SaleResponse struct {
	ID           uuid.UUID `json:"id"`
	Medicine     string    `json:"medicine"`
	PricePerUnit float64   `json:"price_per_unit"`
	Unit         string    `json:"unit"`
	ImageURL     string    `json:"image_url"`
	Quantity     int       `json:"quantity"`
	CreatedAt    time.Time `json:"created_at"`
}
