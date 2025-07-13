package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"pharmacy-management-backend/domain"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// SaleRepository defines the interface for sale-related database operations
type SaleRepository interface {
	SearchMedicines(ctx context.Context, pharmacyID uuid.UUID, query string) ([]domain.MedicineVariant, error)
	AddToCart(ctx context.Context, cart domain.Cart) error
	GetCart(ctx context.Context, userID uuid.UUID) ([]domain.Cart, error)
	RemoveFromCart(ctx context.Context, cartID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
	CreateSale(ctx context.Context, sale domain.Sale, items []domain.SaleItem, receipt domain.Receipt) error
	GetSales(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]domain.Sale, error)
	GetSaleByID(ctx context.Context, saleID uuid.UUID) (*domain.Sale, error)
	GetReceiptBySaleID(ctx context.Context, saleID uuid.UUID) (*domain.Receipt, error)
}

// saleRepository implements SaleRepository
type saleRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewSaleRepository creates a new SaleRepository
func NewSaleRepository(db *sql.DB, logger zerolog.Logger) SaleRepository {
	return &saleRepository{db, logger}
}

// SearchMedicines searches for medicine variants by name or barcode
func (r *saleRepository) SearchMedicines(ctx context.Context, pharmacyID uuid.UUID, query string) ([]domain.MedicineVariant, error) {
	sqlQuery := `
        SELECT mv.id, mv.medicine_id, mv.brand, mv.barcode, mv.unit, mv.price_per_unit, mv.expiry_date, mv.stock, mv.created_at, mv.updated_at
        FROM medicine_variants mv
        JOIN medicines m ON mv.medicine_id = m.id
        WHERE m.pharmacy_id = $1
        AND (m.name ILIKE $2 OR mv.brand ILIKE $2 OR mv.barcode = $3)
        AND mv.expiry_date > NOW()
        AND mv.stock > 0
    `
	rows, err := r.db.QueryContext(ctx, sqlQuery, pharmacyID, "%"+query+"%", query)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to search medicines")
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

// AddToCart adds an item to the cart
func (r *saleRepository) AddToCart(ctx context.Context, cart domain.Cart) error {
	query := `
        INSERT INTO carts (id, user_id, pharmacy_id, medicine_variant_id, quantity, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (user_id, medicine_variant_id)
        DO UPDATE SET quantity = carts.quantity + EXCLUDED.quantity, created_at = EXCLUDED.created_at
    `
	_, err := r.db.ExecContext(ctx, query, cart.ID, cart.UserID, cart.PharmacyID, cart.MedicineVariantID, cart.Quantity, cart.CreatedAt)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to add to cart")
		return err
	}
	return nil
}

// GetCart retrieves cart items for a user
func (r *saleRepository) GetCart(ctx context.Context, userID uuid.UUID) ([]domain.Cart, error) {
	query := `
        SELECT id, user_id, pharmacy_id, medicine_variant_id, quantity, created_at
        FROM carts WHERE user_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get cart")
		return nil, err
	}
	defer rows.Close()

	var carts []domain.Cart
	for rows.Next() {
		var c domain.Cart
		if err := rows.Scan(&c.ID, &c.UserID, &c.PharmacyID, &c.MedicineVariantID, &c.Quantity, &c.CreatedAt); err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan cart item")
			return nil, err
		}
		carts = append(carts, c)
	}
	return carts, nil
}

// RemoveFromCart removes an item from the cart
func (r *saleRepository) RemoveFromCart(ctx context.Context, cartID uuid.UUID) error {
	query := `DELETE FROM carts WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, cartID)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to remove from cart")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check rows affected")
		return err
	}
	if rowsAffected == 0 {
		r.logger.Info().Str("cart_id", cartID.String()).Msg("Cart item not found")
		return domain.ErrCartItemNotFound
	}
	return nil
}

// ClearCart clears all cart items for a user
func (r *saleRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM carts WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to clear cart")
		return err
	}
	return nil
}

// CreateSale creates a sale, sale items, and receipt in a transaction
func (r *saleRepository) CreateSale(ctx context.Context, sale domain.Sale, items []domain.SaleItem, receipt domain.Receipt) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	// Insert sale
	query := `
        INSERT INTO sales (id, user_id, pharmacy_id, total_price, sale_date, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	if _, err := tx.ExecContext(ctx, query, sale.ID, sale.UserID, sale.PharmacyID, sale.TotalPrice, sale.SaleDate, sale.CreatedAt, sale.UpdatedAt); err != nil {
		r.logger.Error().Err(err).Msg("Failed to create sale")
		return err
	}

	// Insert sale items and update stock
	for _, item := range items {
		// Update stock with locking
		updateQuery := `
            UPDATE medicine_variants
            SET stock = stock - $1, updated_at = $2
            WHERE id = $3 AND stock >= $1
        `
		result, err := tx.ExecContext(ctx, updateQuery, item.Quantity, time.Now(), item.MedicineVariantID)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to update stock")
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to check rows affected")
			return err
		}
		if rowsAffected == 0 {
			r.logger.Info().Str("variant_id", item.MedicineVariantID.String()).Msg("Insufficient stock")
			return domain.ErrInsufficientStock
		}

		// Insert sale item
		itemQuery := `
            INSERT INTO sale_items (id, sale_id, medicine_variant_id, quantity, price_per_unit, created_at)
            VALUES ($1, $2, $3, $4, $5, $6)
        `
		if _, err := tx.ExecContext(ctx, itemQuery, item.ID, item.SaleID, item.MedicineVariantID, item.Quantity, item.PricePerUnit, item.CreatedAt); err != nil {
			r.logger.Error().Err(err).Msg("Failed to create sale item")
			return err
		}
	}

	// Insert receipt
	receiptQuery := `
        INSERT INTO receipts (id, sale_id, content, created_at)
        VALUES ($1, $2, $3, $4)
    `
	content, err := json.Marshal(receipt.Content)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to marshal receipt content")
		return err
	}
	if _, err := tx.ExecContext(ctx, receiptQuery, receipt.ID, receipt.SaleID, content, receipt.CreatedAt); err != nil {
		r.logger.Error().Err(err).Msg("Failed to create receipt")
		return err
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error().Err(err).Msg("Failed to commit transaction")
		return err
	}
	return nil
}

// GetSales retrieves sales for a pharmacy
func (r *saleRepository) GetSales(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]domain.Sale, error) {
	query := `
        SELECT id, user_id, pharmacy_id, total_price, sale_date, created_at, updated_at
        FROM sales
        WHERE ($1::uuid IS NULL OR pharmacy_id = $1)
        ORDER BY sale_date DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get sales")
		return nil, err
	}
	defer rows.Close()

	var sales []domain.Sale
	for rows.Next() {
		var s domain.Sale
		if err := rows.Scan(&s.ID, &s.UserID, &s.PharmacyID, &s.TotalPrice, &s.SaleDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan sale")
			return nil, err
		}
		sales = append(sales, s)
	}
	return sales, nil
}

// GetSaleByID retrieves a sale by ID
func (r *saleRepository) GetSaleByID(ctx context.Context, saleID uuid.UUID) (*domain.Sale, error) {
	query := `
        SELECT id, user_id, pharmacy_id, total_price, sale_date, created_at, updated_at
        FROM sales WHERE id = $1
    `
	var s domain.Sale
	err := r.db.QueryRowContext(ctx, query, saleID).Scan(&s.ID, &s.UserID, &s.PharmacyID, &s.TotalPrice, &s.SaleDate, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("sale_id", saleID.String()).Msg("Sale not found")
		return nil, domain.ErrSaleNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get sale by ID")
		return nil, err
	}
	return &s, nil
}

// GetReceiptBySaleID retrieves a receipt by sale ID
func (r *saleRepository) GetReceiptBySaleID(ctx context.Context, saleID uuid.UUID) (*domain.Receipt, error) {
	query := `
        SELECT id, sale_id, content, created_at
        FROM receipts WHERE sale_id = $1
    `
	var receipt domain.Receipt
	var content string
	err := r.db.QueryRowContext(ctx, query, saleID).Scan(&receipt.ID, &receipt.SaleID, &content, &receipt.CreatedAt)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("sale_id", saleID.String()).Msg("Receipt not found")
		return nil, domain.ErrSaleNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get receipt by sale ID")
		return nil, err
	}
	receipt.Content = content
	return &receipt, nil
}
