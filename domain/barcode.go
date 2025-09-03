package domain

import "github.com/google/uuid"

type Barcode struct {
	BarcodeValue string    `json:"barcode_value"`
	MedicineID   uuid.UUID `json:"medicine_id"`
}
