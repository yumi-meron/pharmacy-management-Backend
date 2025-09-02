package domain

import "github.com/google/uuid"

type Barcode struct {
	BarcodeValue      string    `json:"barcode_value"`
	MedicineVariantID uuid.UUID `json:"medicine_variant_id"`
}
