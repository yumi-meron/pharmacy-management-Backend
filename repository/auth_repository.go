package repository

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"pharmacy-management-backend/domain"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	storage "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
)

// AuthRepository defines the interface for authentication-related database operations
type AuthRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByPhone(ctx context.Context, phoneNumber string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user domain.User) (*domain.User, error)
	UploadProfilePicture(ctx context.Context, userID uuid.UUID, fileData []byte, fileExt string) (string, error)
	SaveRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (*uuid.UUID, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	SaveResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetResetToken(ctx context.Context, token string) (*uuid.UUID, error)
	DeleteResetToken(ctx context.Context, token string) error
	GetPharmacists(ctx context.Context, pharmacyID *uuid.UUID) ([]domain.User, error)
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error
}

// authRepository implements AuthRepository
type authRepository struct {
	db       *sql.DB
	logger   zerolog.Logger
	supabase *supabase.Client
	url      string
}

// NewAuthRepository creates a new AuthRepository
func NewAuthRepository(db *sql.DB, logger zerolog.Logger, supabaseClient *supabase.Client, supabaseURL string) AuthRepository {
	return &authRepository{
		db:       db,
		logger:   logger,
		supabase: supabaseClient,
		url:      supabaseURL,
	}
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
func (r *authRepository) Update(ctx context.Context, user domain.User) (*domain.User, error) {
	query := `
		UPDATE users
		SET phone_number = COALESCE(NULLIF($2, ''), phone_number),
			password = COALESCE(NULLIF($3, ''), password),
			full_name = COALESCE(NULLIF($4, ''), full_name),
			role = COALESCE(NULLIF($5, ''), role),
			pharmacy_id = COALESCE(NULLIF($6::uuid, NULL), pharmacy_id),
			profile_picture = COALESCE(NULLIF($7, ''), profile_picture),
			updated_at = $8
		WHERE id = $1
		RETURNING id, phone_number, password, full_name, role, pharmacy_id, profile_picture, created_at, updated_at
	`
	var updatedUser domain.User
	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.PhoneNumber, user.Password, user.FullName, user.Role, user.PharmacyID, user.ProfilePicture, user.UpdatedAt,
	).Scan(
		&updatedUser.ID, &updatedUser.PhoneNumber, &updatedUser.Password, &updatedUser.FullName,
		&updatedUser.Role, &updatedUser.PharmacyID, &updatedUser.ProfilePicture,
		&updatedUser.CreatedAt, &updatedUser.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("id", user.ID.String()).Msg("User not found for update")
		return nil, domain.ErrNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to update user")
		return nil, err
	}
	return &updatedUser, nil
}
func (r *authRepository) UploadProfilePicture(ctx context.Context, userID uuid.UUID, fileData []byte, fileExt string) (string, error) {
	if !strings.HasPrefix(r.url, "https://") {
		r.logger.Error().Str("url", r.url).Msg("Invalid Supabase storage URL")
		return "", fmt.Errorf("invalid Supabase storage URL: %s", r.url)
	}

	if len(fileData) == 0 {
		r.logger.Error().Msg("Empty file data provided")
		return "", fmt.Errorf("empty file data provided")
	}

	// Normalize file extension
	fileExt = strings.ToLower(strings.TrimPrefix(fileExt, "."))

	// Dynamically determine content type
	contentType := http.DetectContentType(fileData)
	if !strings.HasPrefix(contentType, "image/") {
		r.logger.Error().Str("content_type", contentType).Msg("Unsupported file type")
		return "", fmt.Errorf("unsupported file type: %s", contentType)
	}

	filename := fmt.Sprintf("%s_profile_picture_%d.%s", userID, time.Now().UnixNano(), fileExt)
	bucket := "profile_picture"

	// Log file data preview
	filePreview := fileData
	if len(fileData) > 4 {
		filePreview = fileData[:4]
	}

	r.logger.Info().
		Str("bucket", bucket).
		Str("filename", filename).
		Int("file_size", len(fileData)).
		Str("file_ext", fileExt).
		Bytes("file_preview", filePreview).
		Str("supabase_url", r.url).
		Msg("Uploading profile picture")

	r.logger.Debug().
		Str("bucket", bucket).
		Str("filename", filename).
		Int("file_size", len(fileData)).
		Str("content_type", contentType).
		Msg("Attempting to upload profile picture")

	uploadResponse, err := r.supabase.Storage.UploadFile(bucket, filename, bytes.NewReader(fileData), storage.FileOptions{
		ContentType: &contentType,
	})
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("bucket", bucket).
			Str("filename", filename).
			Msgf("Failed to upload profile picture: %v", err)
		return "", fmt.Errorf("failed to upload profile picture to bucket %s: %w", bucket, err)
	}

	// Construct public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s", r.url, uploadResponse.Key)
	r.logger.Info().Str("public_url", publicURL).Msg("Profile picture uploaded successfully")

	// Verify the URL is accessible
	resp, err := http.Head(publicURL)
	if err != nil {
		r.logger.Debug().
			Int("file_size", len(fileData)).
			Str("content_type", contentType).
			Msg("Validating file data")
		r.logger.Error().
			Err(err).
			Str("public_url", publicURL).
			Int("status_code", resp.StatusCode).
			Msg("Failed to verify public URL")
		return "", fmt.Errorf("failed to verify public URL: %v", err)
	}

	return publicURL, nil
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

// GetPharmacists retrieves all pharmacists, optionally filtered by pharmacy_id
func (r *authRepository) GetPharmacists(ctx context.Context, pharmacyID *uuid.UUID) ([]domain.User, error) {
	query := `
        SELECT id, phone_number, password, full_name, role, pharmacy_id, profile_picture, created_at, updated_at
        FROM users
        WHERE role = $1
    `
	args := []interface{}{domain.RolePharmacist}
	if pharmacyID != nil {
		query += ` AND pharmacy_id = $2`
		args = append(args, *pharmacyID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get pharmacists")
		return nil, err
	}
	defer rows.Close()

	var pharmacists []domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID, &user.PhoneNumber, &user.Password, &user.FullName, &user.Role,
			&user.PharmacyID, &user.ProfilePicture, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan pharmacist")
			return nil, err
		}
		pharmacists = append(pharmacists, user)
	}

	return pharmacists, nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *authRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM blacklisted_tokens
            WHERE token = $1 AND expires_at > $2
        )
    `
	var exists bool
	err := r.db.QueryRowContext(ctx, query, token, time.Now()).Scan(&exists)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check if token is blacklisted")
		return false, err
	}
	return exists, nil
}

// BlacklistToken adds a token to the blacklist
func (r *authRepository) BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
	query := `
        INSERT INTO blacklisted_tokens (id, token, expires_at, created_at)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.db.ExecContext(ctx, query, uuid.New(), token, expiresAt, time.Now())
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to blacklist token")
		return err
	}
	return nil
}
