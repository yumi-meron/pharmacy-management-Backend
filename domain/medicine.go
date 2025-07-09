package domain

import (
	"time"

	"github.com/google/uuid"
)

type Medicine struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	Name        string    `json:"name" validate:"required,min=2,max=100"`
	Description string    `json:"description" validate:"max=500"`
	Picture     string    `json:"picture" validate:"omitempty,url"`
	CreatedAt   time.Time `json:"created_at" validate:"required"`
	UpdatedAt   time.Time `json:"updated_at" validate:"required"`
}
