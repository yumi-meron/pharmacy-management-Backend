package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role defines the possible user roles
type Role string

const (
	RolePharmacist Role = "pharmacist"
	RoleOwner      Role = "owner"
	RoleAdmin      Role = "admin"
)

// User represents a user entity in the system
type User struct {
	ID             uuid.UUID `json:"id" validate:"required"`
	PhoneNumber    string    `json:"phone_number" validate:"required,phone" `
	Password       string    `json:"password" validate:"required,min=2"`
	FullName       string    `json:"full_name" validate:"required,min=2,max=100"`
	Role           Role      `json:"role" validate:"required,role"`
	PharmacyID     uuid.UUID `json:"pharmacy_id" validate:"required"`
	ProfilePicture string    `json:"profile_picture" validate:"omitempty,url"` // URL to user profile image
	CreatedAt      time.Time `json:"created_at" validate:"required"`
	UpdatedAt      time.Time `json:"updated_at" validate:"required"`
}

// SignupInput represents the input for user signup
type SignupInput struct {
	PhoneNumber    string    `json:"phone_number" validate:"required,phone" `
	Password       string    `json:"password" validate:"required,min=2"`
	FullName       string    `json:"full_name" validate:"required,min=2,max=100"`
	Role           Role      `json:"role" validate:"required,role"`
	PharmacyID     uuid.UUID `json:"pharmacy_id" validate:"required"`
	ProfilePicture string    `json:"profile_picture" validate:"omitempty,url"`
}
