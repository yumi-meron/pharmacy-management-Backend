package domain

import "github.com/google/uuid"

type SearchParams struct {
	Query      string    `json:"query" validate:"omitempty,min=1,max=100"`
	Filter     string    `json:"filter" validate:"omitempty,oneof=name brand description"` // Default: "name"
	PharmacyID uuid.UUID `json:"pharmacy_id" validate:"omitempty,uuid"`
	Limit      int       `json:"limit" validate:"omitempty,gte=1,lte=100"`
	Offset     int       `json:"offset" validate:"omitempty,gte=0"`
}
