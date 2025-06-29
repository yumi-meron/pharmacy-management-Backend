package repository

import (
	"context"
	"database/sql"

	"github.com/pharmacist-backend/domain"
)

type AuthRepository interface {
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, phone_number, password_hash, full_name, role, pharmacy_id
		FROM users WHERE phone_number = $1`, phone).
		Scan(&u.ID, &u.PhoneNumber, &u.PasswordHash, &u.FullName, &u.Role, &u.PharmacyID)

	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *authRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, phone_number, password_hash, full_name, role, pharmacy_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.PhoneNumber, user.PasswordHash, user.FullName, user.Role, user.PharmacyID)
	return err
}
