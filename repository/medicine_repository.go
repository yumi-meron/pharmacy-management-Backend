package repository

import (
	"context"
	"database/sql"

	"pharmacist-backend/domain"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// MedicineRepository defines the interface for medicine-related database operations
type MedicineRepository interface {
	Create(ctx context.Context, medicine domain.Medicine) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Medicine, error)
	GetAll(ctx context.Context, pharmacyID uuid.UUID) ([]domain.Medicine, error)
	Update(ctx context.Context, medicine domain.Medicine) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountVariants(ctx context.Context, medicineID uuid.UUID) (int, error)
	CreateVariant(ctx context.Context, variant domain.MedicineVariant) error
	GetVariantByID(ctx context.Context, id uuid.UUID) (*domain.MedicineVariant, error)
	GetVariantsByMedicineID(ctx context.Context, medicineID uuid.UUID) ([]domain.MedicineVariant, error)
	UpdateVariant(ctx context.Context, variant domain.MedicineVariant) error
	DeleteVariant(ctx context.Context, id uuid.UUID) error
	CheckBarcodeExists(ctx context.Context, barcode string) (bool, error)
}

// medicineRepository implements MedicineRepository
type medicineRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewMedicineRepository creates a new MedicineRepository
func NewMedicineRepository(db *sql.DB, logger zerolog.Logger) MedicineRepository {
	return &medicineRepository{db, logger}
}

// Create inserts a new medicine into the database
func (r *medicineRepository) Create(ctx context.Context, medicine domain.Medicine) error {
	query := `
        INSERT INTO medicines (id, pharmacy_id, name, description, picture, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.db.ExecContext(ctx, query,
		medicine.ID, medicine.PharmacyID, medicine.Name, medicine.Description, medicine.Picture, medicine.CreatedAt, medicine.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create medicine")
		return err
	}
	return nil
}

// GetByID retrieves a medicine by ID
func (r *medicineRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Medicine, error) {
	query := `
        SELECT id, pharmacy_id, name, description, picture, created_at, updated_at
        FROM medicines WHERE id = $1
    `
	var m domain.Medicine
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.PharmacyID, &m.Name, &m.Description, &m.Picture, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("id", id.String()).Msg("Medicine not found")
		return nil, domain.ErrMedicineNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get medicine by ID")
		return nil, err
	}
	return &m, nil
}

// GetAll retrieves medicines for a pharmacy (or all for Admin)
func (r *medicineRepository) GetAll(ctx context.Context, pharmacyID uuid.UUID) ([]domain.Medicine, error) {
	query := `
        SELECT id, pharmacy_id, name, description, picture, created_at, updated_at
        FROM medicines
        WHERE ($1::uuid IS NULL OR pharmacy_id = $1)
    `
	rows, err := r.db.QueryContext(ctx, query, pharmacyID)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get all medicines")
		return nil, err
	}
	defer rows.Close()

	var medicines []domain.Medicine
	for rows.Next() {
		var m domain.Medicine
		if err := rows.Scan(&m.ID, &m.PharmacyID, &m.Name, &m.Description, &m.Picture, &m.CreatedAt, &m.UpdatedAt); err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan medicine")
			return nil, err
		}
		medicines = append(medicines, m)
	}
	return medicines, nil
}

// Update updates a medicine
func (r *medicineRepository) Update(ctx context.Context, medicine domain.Medicine) error {
	query := `
        UPDATE medicines
        SET name = $2, description = $3, picture = $4, updated_at = $5
        WHERE id = $1
    `
	result, err := r.db.ExecContext(ctx, query,
		medicine.ID, medicine.Name, medicine.Description, medicine.Picture, medicine.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to update medicine")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check rows affected")
		return err
	}
	if rowsAffected == 0 {
		r.logger.Info().Str("id", medicine.ID.String()).Msg("Medicine not found for update")
		return domain.ErrMedicineNotFound
	}
	return nil
}

// Delete deletes a medicine
func (r *medicineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM medicines WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to delete medicine")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check rows affected")
		return err
	}
	if rowsAffected == 0 {
		r.logger.Info().Str("id", id.String()).Msg("Medicine not found for deletion")
		return domain.ErrMedicineNotFound
	}
	return nil
}

// CountVariants counts variants for a medicine
func (r *medicineRepository) CountVariants(ctx context.Context, medicineID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM medicine_variants WHERE medicine_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, medicineID).Scan(&count)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count variants")
		return 0, err
	}
	return count, nil
}

// CreateVariant inserts a new medicine variant
func (r *medicineRepository) CreateVariant(ctx context.Context, variant domain.MedicineVariant) error {
	query := `
        INSERT INTO medicine_variants (id, medicine_id, brand, barcode, unit, price_per_unit, expiry_date, stock, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `
	_, err := r.db.ExecContext(ctx, query,
		variant.ID, variant.MedicineID, variant.Brand, variant.Barcode, variant.Unit, variant.PricePerUnit, variant.ExpiryDate, variant.Stock, variant.CreatedAt, variant.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create medicine variant")
		return err
	}
	return nil
}

// GetVariantByID retrieves a medicine variant by ID
func (r *medicineRepository) GetVariantByID(ctx context.Context, id uuid.UUID) (*domain.MedicineVariant, error) {
	query := `
        SELECT id, medicine_id, brand, barcode, unit, price_per_unit, expiry_date, stock, created_at, updated_at
        FROM medicine_variants WHERE id = $1
    `
	var v domain.MedicineVariant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.MedicineID, &v.Brand, &v.Barcode, &v.Unit, &v.PricePerUnit, &v.ExpiryDate, &v.Stock, &v.CreatedAt, &v.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("id", id.String()).Msg("Medicine variant not found")
		return nil, domain.ErrVariantNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get medicine variant by ID")
		return nil, err
	}
	return &v, nil
}

// GetVariantsByMedicineID retrieves variants for a medicine
func (r *medicineRepository) GetVariantsByMedicineID(ctx context.Context, medicineID uuid.UUID) ([]domain.MedicineVariant, error) {
	query := `
        SELECT id, medicine_id, brand, barcode, unit, price_per_unit, expiry_date, stock, created_at, updated_at
        FROM medicine_variants WHERE medicine_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, medicineID)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get medicine variants")
		return nil, err
	}
	defer rows.Close()

	var variants []domain.MedicineVariant
	for rows.Next() {
		var v domain.MedicineVariant
		if err := rows.Scan(&v.ID, &v.MedicineID, &v.Brand, &v.Barcode, &v.Unit, &v.PricePerUnit, &v.ExpiryDate, &v.Stock, &v.CreatedAt, &v.UpdatedAt); err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan medicine variant")
			return nil, err
		}
		variants = append(variants, v)
	}
	return variants, nil
}

// UpdateVariant updates a medicine variant
func (r *medicineRepository) UpdateVariant(ctx context.Context, variant domain.MedicineVariant) error {
	query := `
        UPDATE medicine_variants
        SET brand = $2, barcode = $3, unit = $4, price_per_unit = $5, expiry_date = $6, stock = $7, updated_at = $8
        WHERE id = $1
    `
	result, err := r.db.ExecContext(ctx, query,
		variant.ID, variant.Brand, variant.Barcode, variant.Unit, variant.PricePerUnit, variant.ExpiryDate, variant.Stock, variant.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to update medicine variant")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check rows affected")
		return err
	}
	if rowsAffected == 0 {
		r.logger.Info().Str("id", variant.ID.String()).Msg("Medicine variant not found for update")
		return domain.ErrVariantNotFound
	}
	return nil
}

// DeleteVariant deletes a medicine variant
func (r *medicineRepository) DeleteVariant(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM medicine_variants WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to delete medicine variant")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check rows affected")
		return err
	}
	if rowsAffected == 0 {
		r.logger.Info().Str("id", id.String()).Msg("Medicine variant not found for deletion")
		return domain.ErrVariantNotFound
	}
	return nil
}

// CheckBarcodeExists checks if a barcode is already taken
func (r *medicineRepository) CheckBarcodeExists(ctx context.Context, barcode string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM medicine_variants WHERE barcode = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, barcode).Scan(&exists)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check barcode existence")
		return false, err
	}
	return exists, nil
}
