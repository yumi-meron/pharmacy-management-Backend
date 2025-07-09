package repository

import (
	"context"
	"database/sql"

	"pharmacist-backend/domain"

	"github.com/google/uuid"
)

// PharmacyRepository defines the interface for pharmacy-related database operations
type PharmacyRepository interface {
	Create(ctx context.Context, pharmacy domain.Pharmacy) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Pharmacy, error)
	GetAll(ctx context.Context, offset, limit int) ([]domain.Pharmacy, error)
	Update(ctx context.Context, pharmacy domain.Pharmacy) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// pharmacyRepository implements PharmacyRepository
type pharmacyRepository struct {
	db *sql.DB
}

// NewPharmacyRepository creates a new PharmacyRepository
func NewPharmacyRepository(db *sql.DB) PharmacyRepository {
	return &pharmacyRepository{db: db}
}

// Create inserts a new pharmacy into the database
func (r *pharmacyRepository) Create(ctx context.Context, pharmacy domain.Pharmacy) error {
	query := `
        INSERT INTO pharmacies (id, name, address, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query,
		pharmacy.ID,
		pharmacy.Name,
		pharmacy.Address,
		pharmacy.CreatedAt,
		pharmacy.UpdatedAt,
	)
	return err
}

// GetByID retrieves a pharmacy by ID
func (r *pharmacyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pharmacy, error) {
	query := `
        SELECT id, name, address, created_at, updated_at
        FROM pharmacies WHERE id = $1
    `
	var pharmacy domain.Pharmacy
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pharmacy.ID,
		&pharmacy.Name,
		&pharmacy.Address,
		&pharmacy.CreatedAt,
		&pharmacy.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPharmacyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &pharmacy, nil
}

// GetAll retrieves all pharmacies with pagination
func (r *pharmacyRepository) GetAll(ctx context.Context, offset, limit int) ([]domain.Pharmacy, error) {
	query := `
        SELECT id, name, address, created_at, updated_at
        FROM pharmacies
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pharmacies []domain.Pharmacy
	for rows.Next() {
		var pharmacy domain.Pharmacy
		if err := rows.Scan(
			&pharmacy.ID,
			&pharmacy.Name,
			&pharmacy.Address,
			&pharmacy.CreatedAt,
			&pharmacy.UpdatedAt,
		); err != nil {
			return nil, err
		}
		pharmacies = append(pharmacies, pharmacy)
	}
	return pharmacies, nil
}

// Update updates a pharmacy's details
func (r *pharmacyRepository) Update(ctx context.Context, pharmacy domain.Pharmacy) error {
	query := `
        UPDATE pharmacies
        SET name = $2, address = $3, updated_at = $4
        WHERE id = $1
    `
	result, err := r.db.ExecContext(ctx, query,
		pharmacy.ID,
		pharmacy.Name,
		pharmacy.Address,
		pharmacy.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrPharmacyNotFound
	}
	return nil
}

// Delete removes a pharmacy by ID
func (r *pharmacyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM pharmacies WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrPharmacyNotFound
	}
	return nil
}
