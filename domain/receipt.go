package domain

import (
	"time"

	"github.com/google/uuid"
)

type Receipt struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	SaleID    uuid.UUID `json:"sale_id" validate:"required"`
	Content   string    `json:"content" validate:"required"` // JSON string
	CreatedAt time.Time `json:"created_at" validate:"required"`
}
