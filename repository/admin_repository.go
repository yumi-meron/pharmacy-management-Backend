package repository

import (
	"context"
	"database/sql"

	"pharmacist-backend/domain"
)

// AdminRepository defines the interface for admin-related database operations
type AdminRepository interface {
	CreateOwner(ctx context.Context, user domain.User) error
}

// adminRepository implements AdminRepository
type adminRepository struct {
	db *sql.DB
}

// NewAdminRepository creates a new AdminRepository
func NewAdminRepository(db *sql.DB) AdminRepository {
	return &adminRepository{db: db}
}

// CreateOwner creates a new user with the owner role
func (r *adminRepository) CreateOwner(ctx context.Context, user domain.User) error {
	query := `
        INSERT INTO users (id, phone_number, password, full_name, role, pharmacy_id, profile_picture, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.PhoneNumber,
		user.Password, // Assumes password is already hashed
		user.FullName,
		user.Role,
		user.PharmacyID,
		user.ProfilePicture,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_phone_number_key\"" {
			return domain.ErrPhoneNumberTaken
		}
		return err
	}
	return nil
}
