package repository

import (
	"context"
	"database/sql"
	"errors"
	"pharmacy-management-backend/domain"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// BarcodeRepository defines the interface for barcode-related database operations
type BarcodeRepository interface {
	LookupBarcode(ctx context.Context, barcodeValue string) (uuid.UUID, error)
	AddBarcodes(ctx context.Context, barcodes domain.Barcode) error
}

// barcodeRepository implements BarcodeRepository
type barcodeRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewBarcodeRepository creates a new BarcodeRepository
func NewBarcodeRepository(db *sql.DB, logger zerolog.Logger) BarcodeRepository {
	return &barcodeRepository{db: db, logger: logger}
}

// LookupBarcode looks up a barcode in the database and returns the associated medicine ID
func (r *barcodeRepository) LookupBarcode(ctx context.Context, barcodeValue string) (uuid.UUID, error) {
	var medicineID uuid.UUID
	query := `SELECT medicine_id FROM barcodes WHERE barcode_value = $1`
	err := r.db.QueryRowContext(ctx, query, barcodeValue).Scan(&medicineID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, domain.ErrBarcodeNotFound
		}
		r.logger.Error().Err(err).Msg("failed to lookup barcode")
		return uuid.Nil, err
	}
	return medicineID, nil
}

// CreateBarcode creates a new barcode entry in the database
func (r *barcodeRepository) AddBarcodes(ctx context.Context, barcodes domain.Barcode) error {
	query := `INSERT INTO barcodes (barcode_value, medicine_id) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, barcodes.BarcodeValue, barcodes.MedicineID)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to add barcode")
		return err
	}

	return nil
}
