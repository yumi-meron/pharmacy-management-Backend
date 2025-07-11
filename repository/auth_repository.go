package repository

import (
	"context"
	"database/sql"
	"time"

	"pharmacy-management-backend/domain"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AuthRepository defines the interface for authentication-related database operations
type AuthRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByPhone(ctx context.Context, phoneNumber string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user domain.User) error
	SaveRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (*uuid.UUID, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	SaveResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetResetToken(ctx context.Context, token string) (*uuid.UUID, error)
	DeleteResetToken(ctx context.Context, token string) error
}

// authRepository implements AuthRepository
type authRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewAuthRepository creates a new AuthRepository
func NewAuthRepository(db *sql.DB, logger zerolog.Logger) AuthRepository {
	return &authRepository{db, logger}
}

// Create inserts a new user into the database
func (r *authRepository) Create(ctx context.Context, user domain.User) error {
	query := `
        INSERT INTO users (id, phone_number, password, full_name, role, pharmacy_id, profile_picture, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.PhoneNumber, user.Password, user.FullName, user.Role, user.PharmacyID, user.ProfilePicture, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create user")
		return err
	}
	return nil
}

// GetByPhone retrieves a user by phone number
func (r *authRepository) GetByPhone(ctx context.Context, phoneNumber string) (*domain.User, error) {
	query := `
        SELECT id, phone_number, password, full_name, role, pharmacy_id, profile_picture, created_at, updated_at
        FROM users WHERE phone_number = $1
    `
	var user domain.User
	err := r.db.QueryRowContext(ctx, query, phoneNumber).Scan(
		&user.ID, &user.PhoneNumber, &user.Password, &user.FullName, &user.Role, &user.PharmacyID, &user.ProfilePicture, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("phone_number", phoneNumber).Msg("User not found")
		return nil, domain.ErrNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get user by phone")
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by ID
func (r *authRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
        SELECT id, phone_number, password, full_name, role, pharmacy_id, profile_picture, created_at, updated_at
        FROM users WHERE id = $1
    `
	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.PhoneNumber, &user.Password, &user.FullName, &user.Role, &user.PharmacyID, &user.ProfilePicture, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("id", id.String()).Msg("User not found")
		return nil, domain.ErrNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get user by ID")
		return nil, err
	}
	return &user, nil
}

// Update updates a user in the database
func (r *authRepository) Update(ctx context.Context, user domain.User) error {
	query := `
        UPDATE users
        SET phone_number = $2, password = $3, full_name = $4, role = $5, pharmacy_id = $6, profile_picture = $7, updated_at = $8
        WHERE id = $1
    `
	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.PhoneNumber, user.Password, user.FullName, user.Role, user.PharmacyID, user.ProfilePicture, user.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to update user")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check rows affected")
		return err
	}
	if rowsAffected == 0 {
		r.logger.Info().Str("id", user.ID.String()).Msg("User not found for update")
		return domain.ErrNotFound
	}
	return nil
}

// SaveRefreshToken saves a refresh token
func (r *authRepository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
        INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $5)
    `
	_, err := r.db.ExecContext(ctx, query, uuid.New(), userID, token, expiresAt, time.Now())
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to save refresh token")
		return err
	}
	return nil
}

// GetRefreshToken retrieves a user ID by refresh token
func (r *authRepository) GetRefreshToken(ctx context.Context, token string) (*uuid.UUID, error) {
	query := `
        SELECT user_id
        FROM refresh_tokens
        WHERE token = $1 AND expires_at > $2
    `
	var userID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, token, time.Now()).Scan(&userID)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("token", token).Msg("Refresh token not found or expired")
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get refresh token")
		return nil, err
	}
	return &userID, nil
}

// DeleteRefreshToken deletes a refresh token
func (r *authRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to delete refresh token")
		return err
	}
	return nil
}

// SaveResetToken saves a password reset token
func (r *authRepository) SaveResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
        INSERT INTO password_reset_tokens (user_id, token, expires_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $4)
    `
	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt, time.Now())
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to save reset token")
		return err
	}
	return nil
}

// GetResetToken retrieves a user ID by reset token
func (r *authRepository) GetResetToken(ctx context.Context, token string) (*uuid.UUID, error) {
	query := `
        SELECT user_id
        FROM password_reset_tokens
        WHERE token = $1 AND expires_at > $2
    `
	var userID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, token, time.Now()).Scan(&userID)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("token", token).Msg("Reset token not found or expired")
		return nil, domain.ErrInvalidResetToken
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get reset token")
		return nil, err
	}
	return &userID, nil
}

// DeleteResetToken deletes a reset token
func (r *authRepository) DeleteResetToken(ctx context.Context, token string) error {
	query := `DELETE FROM password_reset_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to delete reset token")
		return err
	}
	return nil
}
