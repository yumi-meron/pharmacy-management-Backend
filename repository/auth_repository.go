package repository

import (
	"context"
	"database/sql"

	"github.com/yumi-meron/pharmacy-management-app/pharmacist-backend/module/domain"
)

type AuthRepository interface {
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, phone_number, password_hash, full_name, role, pharmacy_id
		FROM users WHERE phone_number = $1
	`, phone)

	var u domain.User
	err := row.Scan(&u.ID, &u.PhoneNumber, &u.PasswordHash, &u.FullName, &u.Role, &u.PharmacyID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
