package domain

import "github.com/google/uuid"

type Role string

const (
	RoleOwner      Role = "owner"
	RolePharmacist Role = "pharmacist"
)

type User struct {
	ID           uuid.UUID
	PhoneNumber  string
	PasswordHash string
	FullName     string
	Role         Role
	PharmacyID   uuid.UUID
}
