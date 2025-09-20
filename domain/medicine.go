package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Medicine represents a medicine entity
type Medicine struct {
	ID           uuid.UUID         `json:"id" validate:"required"`
	PharmacyID   uuid.UUID         `json:"pharmacy_id" validate:"required"`
	Name         string            `json:"name" validate:"required,min=2,max=100"`
	Description  string            `json:"description" validate:"max=500"`
	MedicalUsage string            `json:"medical_usecase" validate:"max=1000"`
	Picture      string            `json:"picture" validate:"omitempty,url"`
	CreatedAt    time.Time         `json:"created_at" validate:"required"`
	UpdatedAt    time.Time         `json:"updated_at" validate:"required"`
	Variants     []MedicineVariant `json:"variants" validate:"dive"`
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
	Name         string    `json:"name" validate:"required,min=2,max=100"`
	Description  string    `json:"description" validate:"max=500"`
	MedicalUsage string    `json:"medical_usecase" validate:"max=1000"`
	Picture      string    `json:"picture" validate:"omitempty,url"`
	Brand        string    `json:"brand" validate:"required,min=2,max=100"`
	Barcode      string    `json:"barcode" validate:"required,barcode"`
	Unit         string    `json:"unit" validate:"required,min=1,max=50"`
	PricePerUnit float64   `json:"price_per_unit" validate:"required,gt=0"`
	ExpiryDate   time.Time `json:"expiry_date" validate:"required,future_date"`
	Stock        int       `json:"stock" validate:"required,gte=0"`
}

func (v *MedicineVariant) UnmarshalJSON(data []byte) error {
	type Alias MedicineVariant
	aux := &struct {
		ExpiryDate string `json:"expiry_date"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
		*Alias
	}{
		Alias: (*Alias)(v),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse ExpiryDate
	if aux.ExpiryDate != "" {
		parsedTime, err := parseTime(aux.ExpiryDate)
		if err != nil {
			return err
		}
		v.ExpiryDate = parsedTime
	}

	// Parse CreatedAt
	if aux.CreatedAt != "" {
		parsedTime, err := parseTime(aux.CreatedAt)
		if err != nil {
			return err
		}
		v.CreatedAt = parsedTime
	}

	// Parse UpdatedAt
	if aux.UpdatedAt != "" {
		parsedTime, err := parseTime(aux.UpdatedAt)
		if err != nil {
			return err
		}
		v.UpdatedAt = parsedTime
	}

	return nil
}

// Helper function to parse time with or without timezone
func parseTime(value string) (time.Time, error) {
	// Try parsing with RFC3339
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return parsedTime, nil
	}

	// Try parsing without timezone
	parsedTime, err = time.Parse("2006-01-02T15:04:05.999999", value)
	if err == nil {
		return parsedTime, nil
	}

	return time.Time{}, err
}
