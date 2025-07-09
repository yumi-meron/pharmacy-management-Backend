package domain

import (
	"time"

	"github.com/google/uuid"
)

type MedicineVariant struct {
	ID           uuid.UUID `json:"id"`
	MedicineID   uuid.UUID `json:"medicine_id"`
	Brand        string    `json:"brand"`
	Barcode      string    `json:"barcode"`
	Unit         string    `json:"unit"`
	PricePerUnit float64   `json:"price_per_unit"`
	ExpiryDate   time.Time `json:"expiry_date"`
	Stock        int       `json:"quantity_available"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
