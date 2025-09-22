package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"pharmacy-management-backend/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	storage "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
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
	SearchMedicines(ctx context.Context, params domain.SearchParams) ([]domain.Medicine, error)
	UploadPicture(ctx context.Context, medicineID uuid.UUID, fileData []byte, fileExt string) (string, error)
}

// medicineRepository implements MedicineRepository
type medicineRepository struct {
	db       *sql.DB
	logger   zerolog.Logger
	supabase *supabase.Client
	url      string
}

// NewMedicineRepository creates a new MedicineRepository
func NewMedicineRepository(db *sql.DB, logger zerolog.Logger, supabaseClient *supabase.Client, supabaseURL string) MedicineRepository {
	return &medicineRepository{
		db:       db,
		logger:   logger,
		supabase: supabaseClient,
		url:      supabaseURL,
	}
}

// Create inserts a new medicine into the database
func (r *medicineRepository) Create(ctx context.Context, medicine domain.Medicine) error {
	query := `
        INSERT INTO medicines (id, pharmacy_id, name, description, medical_usage, picture, created_at, updated_at)
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
        SELECT id, pharmacy_id, name, description, medical_usage, picture, created_at, updated_at
        FROM medicines WHERE id = $1
    `
	var m domain.Medicine
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.PharmacyID, &m.Name, &m.Description, &m.MedicalUsage, &m.Picture, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("id", id.String()).Msg("Medicine not found")
		return nil, domain.ErrMedicineNotFound
	}
	m.Variants, err = r.GetVariantsByMedicineID(ctx, id)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get medicine variants")
		return nil, err
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
        SELECT id, pharmacy_id, name, description, medical_usage, picture, created_at, updated_at
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
		if err := rows.Scan(&m.ID, &m.PharmacyID, &m.Name, &m.Description, &m.MedicalUsage, &m.Picture, &m.CreatedAt, &m.UpdatedAt); err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan medicine")
			return nil, err
		}
		m.Variants, err = r.GetVariantsByMedicineID(ctx, m.ID)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to get medicine variants")
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
        SET name = $2, description = $3, picture = $4, updated_at = $5, medical_usage = $6
        WHERE id = $1
    `
	result, err := r.db.ExecContext(ctx, query,
		medicine.ID, medicine.Name, medicine.Description, medicine.Picture, medicine.UpdatedAt, medicine.MedicalUsage,
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

func (r *medicineRepository) UploadPicture(ctx context.Context, medicineID uuid.UUID, fileData []byte, fileExt string) (string, error) {
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

	filename := fmt.Sprintf("%s_medicine_picture_%d.%s", medicineID, time.Now().UnixNano(), fileExt)
	bucket := "medicine_picture"

	uploadResponse, err := r.supabase.Storage.UploadFile(bucket, filename, bytes.NewReader(fileData), storage.FileOptions{
		ContentType: &contentType,
	})
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("bucket", bucket).
			Str("filename", filename).
			Msgf("Failed to upload medicine picture: %v", err)
		return "", fmt.Errorf("failed to upload medicine picture to bucket %s: %w", bucket, err)
	}

	// Construct public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s", r.url, uploadResponse.Key)
	r.logger.Info().Str("public_url", publicURL).Msg("medicine picture uploaded successfully")

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
func (r *medicineRepository) SearchMedicines(ctx context.Context, params domain.SearchParams) ([]domain.Medicine, error) {
	var query strings.Builder
	var args []interface{}
	argIndex := 1

	query.WriteString(`
	SELECT 
		m.id, m.pharmacy_id, m.name, m.description, m.picture, m.created_at, m.updated_at,
		array_agg(row_to_json(v)::text) as variants_json
	FROM medicines m
	LEFT JOIN medicine_variants v ON v.medicine_id = m.id
	LEFT JOIN categories c ON c.id = m.category_id
	WHERE 1=1
`)

	// PharmacyID filter
	if params.PharmacyID != uuid.Nil {
		query.WriteString(fmt.Sprintf(" AND m.pharmacy_id = $%d", argIndex))
		args = append(args, params.PharmacyID)
		r.logger.Debug().Interface("pharmacy_id", params.PharmacyID).Msg("Applied pharmacy filter")
		argIndex++
	} else {
		r.logger.Debug().Msg("No pharmacy filter applied")
	}

	// Category filter (only include if not "None")
	if params.Catagory != "None" {
		query.WriteString(fmt.Sprintf(" AND c.name ILIKE $%d", argIndex))
		args = append(args, params.Catagory)
		r.logger.Debug().Str("category", params.Catagory).Msg("Applied category filter")
		argIndex++
	} else {
		r.logger.Debug().Str("category", params.Catagory).Msg("Category filter skipped")
	}

	// Query + Filter (full-text search)
	if params.Query != "" {
		var tsField string
		switch params.Filter {
		case "brand":
			tsField = "v.ts_brand"
		case "description":
			tsField = "m.ts_description"
		default: // "name"
			tsField = "m.ts_name"
		}
		query.WriteString(fmt.Sprintf(" AND %s @@ to_tsquery('english', $%d)", tsField, argIndex))
		args = append(args, params.Query)
		r.logger.Debug().
			Str("filter", params.Filter).
			Str("query", params.Query).
			Str("ts_field", tsField).
			Msg("Applied text search filter")
		argIndex++
	} else {
		r.logger.Debug().Msg("No text search filter applied")
	}

	// Group and paginate
	query.WriteString(fmt.Sprintf(`
	GROUP BY m.id, m.pharmacy_id, m.name, m.description, m.picture, m.created_at, m.updated_at
	ORDER BY m.name ASC
	LIMIT $%d OFFSET $%d
`, argIndex, argIndex+1))
	args = append(args, params.Limit, params.Offset)

	r.logger.Debug().
		Str("final_query", query.String()).
		Interface("args", args).
		Msg("Executing SearchMedicines query")

	// Execute query
	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to execute search query")
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var medicines []domain.Medicine
	for rows.Next() {
		var m domain.Medicine
		var variantsJSON []string
		err := rows.Scan(
			&m.ID, &m.PharmacyID, &m.Name, &m.Description, &m.Picture, &m.CreatedAt, &m.UpdatedAt,
			pq.Array(&variantsJSON),
		)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan medicine row")
			return nil, fmt.Errorf("failed to scan medicine row: %w", err)
		}

		// Parse variants from JSON
		for _, vj := range variantsJSON {
			if vj == "" {
				continue
			}
			var v domain.MedicineVariant
			if err := json.Unmarshal([]byte(vj), &v); err != nil {
				r.logger.Error().
					Err(err).
					Str("variant_json", vj).
					Msg("Failed to unmarshal variant JSON")
				return nil, fmt.Errorf("failed to unmarshal variant JSON: %w", err)
			}
			m.Variants = append(m.Variants, v)
		}

		r.logger.Debug().
			Str("medicine_id", m.ID.String()).
			Int("variant_count", len(m.Variants)).
			Msg("Fetched medicine row")

		medicines = append(medicines, m)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error().Err(err).Msg("Error iterating medicine rows")
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	r.logger.Debug().Int("result_count", len(medicines)).Msg("SearchMedicines completed successfully")
	return medicines, nil
}
