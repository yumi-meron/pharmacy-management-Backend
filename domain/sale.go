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
}

// CreateCartInput for adding an item to the cart
type CreateCartInput struct {
	MedicineVariantID uuid.UUID `json:"medicine_variant_id" validate:"required"`
	Quantity          int       `json:"quantity" validate:"required,gt=0"`
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

// SaleItem represents an item in a sale
type SaleItem struct {
	ID                uuid.UUID `json:"id" validate:"required"`
	SaleID            uuid.UUID `json:"sale_id" validate:"required"`
	MedicineVariantID uuid.UUID `json:"medicine_variant_id" validate:"required"`
	Quantity          int       `json:"quantity" validate:"required,gt=0"`
	PricePerUnit      float64   `json:"price_per_unit" validate:"required,gt=0"`
	CreatedAt         time.Time `json:"created_at" validate:"required"`
}
