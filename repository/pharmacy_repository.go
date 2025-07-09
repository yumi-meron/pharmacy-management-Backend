package repository

import (
	"context"
	"database/sql"

	"pharmacist-backend/domain"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// PharmacyRepository defines the interface for pharmacy-related database operations
type PharmacyRepository interface {
	Create(ctx context.Context, pharmacy domain.Pharmacy) error
	GetAll(ctx context.Context) ([]domain.Pharmacy, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Pharmacy, error)
	Update(ctx context.Context, pharmacy domain.Pharmacy) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// pharmacyRepository implements PharmacyRepository
type pharmacyRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewPharmacyRepository creates a new PharmacyRepository
func NewPharmacyRepository(db *sql.DB, logger zerolog.Logger) PharmacyRepository {
	return &pharmacyRepository{db, logger}
}

// Create inserts a new pharmacy into the database
func (r *pharmacyRepository) Create(ctx context.Context, pharmacy domain.Pharmacy) error {
	query := `
        INSERT INTO pharmacies (id, name, address, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query,
		pharmacy.ID, pharmacy.Name, pharmacy.Address, pharmacy.CreatedAt, pharmacy.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create pharmacy")
		return err
	}
	return nil
}

// GetAll retrieves all pharmacies
func (r *pharmacyRepository) GetAll(ctx context.Context) ([]domain.Pharmacy, error) {
	query := `
        SELECT id, name, address, created_at, updated_at
        FROM pharmacies
    `
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get all pharmacies")
		return nil, err
	}
	defer rows.Close()

	var pharmacies []domain.Pharmacy
	for rows.Next() {
		var p domain.Pharmacy
		if err := rows.Scan(&p.ID, &p.Name, &p.Address, &p.CreatedAt, &p.UpdatedAt); err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan pharmacy")
			return nil, err
		}
		pharmacies = append(pharmacies, p)
	}
	return pharmacies, nil
}

// GetByID retrieves a pharmacy by ID
func (r *pharmacyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pharmacy, error) {
	query := `
        SELECT id, name, address, created_at, updated_at
        FROM pharmacies WHERE id = $1
    `
	var p domain.Pharmacy
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Address, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("id", id.String()).Msg("Pharmacy not found")
		return nil, domain.ErrNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get pharmacy by ID")
		return nil, err
	}
	return &p, nil
}

// Update updates a pharmacy
func (r *pharmacyRepository) Update(ctx context.Context, pharmacy domain.Pharmacy) error {
	query := `
        UPDATE pharmacies
        SET name = $2, address = $3, updated_at = $4
        WHERE id = $1
    `
	result, err := r.db.ExecContext(ctx, query,
		pharmacy.ID, pharmacy.Name, pharmacy.Address, pharmacy.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to update pharmacy")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check rows affected")
		return err
	}
	if rowsAffected == 0 {
		r.logger.Info().Str("id", pharmacy.ID.String()).Msg("Pharmacy not found for update")
		return domain.ErrNotFound
	}
	return nil
}

// Delete deletes a pharmacy
func (r *pharmacyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM pharmacies WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to delete pharmacy")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check rows affected")
		return err
	}
	if rowsAffected == 0 {
		r.logger.Info().Str("id", id.String()).Msg("Pharmacy not found for deletion")
		return domain.ErrNotFound
	}
	return nil
}
