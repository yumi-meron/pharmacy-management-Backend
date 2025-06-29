package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pharmacist-backend/domain"
)

type AuthRepository interface {
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	SaveOTP(ctx context.Context, phone, otp string, expiresAt time.Time) error
	VerifyOTP(ctx context.Context, phone, otp string) (bool, error)
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
		SELECT id, phone_number, password_hash,

 full_name, role, pharmacy_id
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

func (r *authRepository) SaveOTP(ctx context.Context, phone, otp string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO otp_codes (phone_number, otp, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (phone_number)
		DO UPDATE SET otp = $2, expires_at = $3
	`, phone, otp, expiresAt)
	return err
}

func (r *authRepository) VerifyOTP(ctx context.Context, phone, otp string) (bool, error) {
	var storedOTP string
	var expiresAt time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT otp, expires_at FROM otp_codes WHERE phone_number = $1
	`, phone).Scan(&storedOTP, &expiresAt)
	if err != nil {
		return false, err
	}
	if time.Now().After(expiresAt) {
		return false, domain.ErrInvalidOTP
	}
	return storedOTP == otp, nil
}
