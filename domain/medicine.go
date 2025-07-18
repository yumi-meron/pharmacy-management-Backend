package domain

import (
	"time"

	"github.com/google/uuid"
)

// Medicine represents a medicine entity
type Medicine struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	PharmacyID  uuid.UUID `json:"pharmacy_id" validate:"required"`
	Name        string    `json:"name" validate:"required,min=2,max=100"`
	Description string    `json:"description" validate:"max=500"`
	Picture     string    `json:"picture" validate:"omitempty,url"`
	CreatedAt   time.Time `json:"created_at" validate:"required"`
	UpdatedAt   time.Time `json:"updated_at" validate:"required"`
	Variants    []MedicineVariant `json:"variants" validate:"dive"`

}

// MedicineVariant represents a variant of a medicine
type MedicineVariant struct {
	ID           uuid.UUID `json:"id" validate:"required"`
	MedicineID   uuid.UUID `json:"medicine_id" validate:"required"`
	Brand        string    `json:"brand" validate:"required,min=2,max=100"`
	Barcode      string    `json:"barcode" validate:"required,barcode"`
	Unit         string    `json:"unit" validate:"required,min=1,max=50"`
	PricePerUnit float64   `json:"price_per_unit" validate:"required,gt=0"`
	ExpiryDate   time.Time `json:"expiry_date" validate:"required,future_date"`
	Stock        int       `json:"stock" validate:"required,gte=0"`
	CreatedAt    time.Time `json:"created_at" validate:"required"`
	UpdatedAt    time.Time `json:"updated_at" validate:"required"`
}

// CreateMedicineInput for creating a medicine
type CreateMedicineInput struct {
	PharmacyID  uuid.UUID `json:"pharmacy_id" validate:"required"`
	Name        string    `json:"name" validate:"required,min=2,max=100"`
	Description string    `json:"description" validate:"max=500"`
	Picture     string    `json:"picture" validate:"omitempty,url"`
}

// UpdateMedicineInput for updating a medicine
type UpdateMedicineInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
	Picture     string `json:"picture" validate:"omitempty,url"`
}

// CreateMedicineVariantInput for creating a medicine variant
type CreateMedicineVariantInput struct {
	Brand        string    `json:"brand" validate:"required,min=2,max=100"`
	Barcode      string    `json:"barcode" validate:"required,barcode"`
	Unit         string    `json:"unit" validate:"required,min=1,max=50"`
	PricePerUnit float64   `json:"price_per_unit" validate:"required,gt=0"`
	ExpiryDate   time.Time `json:"expiry_date" validate:"required,future_date"`
	Stock        int       `json:"stock" validate:"required,gte=0"`
}

// UpdateMedicineVariantInput for updating a medicine variant
type UpdateMedicineVariantInput struct {
	Brand        string    `json:"brand" validate:"required,min=2,max=100"`
	Barcode      string    `json:"barcode" validate:"required,barcode"`
	Unit         string    `json:"unit" validate:"required,min=1,max=50"`
	PricePerUnit float64   `json:"price_per_unit" validate:"required,gt=0"`
	ExpiryDate   time.Time `json:"expiry_date" validate:"required,future_date"`
	Stock        int       `json:"stock" validate:"required,gte=0"`
}
