package repository

import (
	"context"
	"database/sql"
	"pharmacy-management-backend/domain"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// OrderRepository defines the interface for order-related database operations
type OrderRepository interface {
	ListOrders(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]domain.OrderResponse, error)
	GetOrderDetails(ctx context.Context, orderID uuid.UUID) (*domain.Order, []domain.OrderItem, *domain.Patient, error)
	GetPatientByID(ctx context.Context, patientID uuid.UUID) (*domain.Patient, error)
	CreateOrder(ctx context.Context, order *domain.Order, items []domain.OrderItem) error
	RequestOTP(ctx context.Context, otp *domain.OrderOTP) error
	VerifyOrder(ctx context.Context, orderID uuid.UUID, otp string) (bool, error)
}

// orderRepository implements OrderRepository
type orderRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewOrderRepository creates a new OrderRepository
func NewOrderRepository(db *sql.DB, logger zerolog.Logger) OrderRepository {
	return &orderRepository{db, logger}
}

// ListOrders retrieves orders for a pharmacy with hospital and patient names
func (r *orderRepository) ListOrders(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]domain.OrderResponse, error) {
	query := `
        SELECT o.id, h.name, p.full_name, o.order_date, o.status
        FROM orders o
        JOIN hospitals h ON o.hospital_id = h.id
        JOIN patients p ON o.patient_id = p.id
        WHERE ($1::uuid IS NULL OR o.pharmacy_id = $1)
        ORDER BY o.order_date DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, query, pharmacyID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to list orders")
		return nil, err
	}
	defer rows.Close()

	var orders []domain.OrderResponse
	for rows.Next() {
		var order domain.OrderResponse
		if err := rows.Scan(&order.ID, &order.HospitalName, &order.PatientName, &order.OrderDate, &order.Status); err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan order")
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// GetOrderDetails retrieves order details including patient and items
func (r *orderRepository) GetOrderDetails(ctx context.Context, orderID uuid.UUID) (*domain.Order, []domain.OrderItem, *domain.Patient, error) {
	query := `
        SELECT o.id, o.hospital_id, o.patient_id, o.pharmacy_id, o.order_date, o.status, o.created_at, o.updated_at
        FROM orders o
        WHERE o.id = $1
    `
	var order domain.Order
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.HospitalID, &order.PatientID, &order.PharmacyID, &order.OrderDate, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("order_id", orderID.String()).Msg("Order not found")
		return nil, nil, nil, domain.ErrOrderNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get order")
		return nil, nil, nil, err
	}

	patientQuery := `
        SELECT id, full_name, phone_number, emergency_phone_number, created_at, updated_at
        FROM patients
        WHERE id = $1
    `
	var patient domain.Patient
	err = r.db.QueryRowContext(ctx, patientQuery, order.PatientID).Scan(
		&patient.ID, &patient.FullName, &patient.PhoneNumber, &patient.EmergencyPhoneNumber, &patient.CreatedAt, &patient.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get patient")
		return nil, nil, nil, err
	}

	itemsQuery := `
        SELECT oi.id, oi.order_id, oi.medicine_variant_id, oi.quantity, oi.price_per_unit, oi.created_at,
               m.name, mv.unit
        FROM order_items oi
        JOIN medicine_variants mv ON oi.medicine_variant_id = mv.id
        JOIN medicines m ON mv.medicine_id = m.id
        WHERE oi.order_id = $1
    `
	rows, err := r.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get order items")
		return nil, nil, nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		var medicineName, unit string
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.MedicineVariantID, &item.Quantity, &item.PricePerUnit, &item.CreatedAt,
			&medicineName, &unit,
		); err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan order item")
			return nil, nil, nil, err
		}
		item.MedicineName = medicineName
		item.Unit = unit
		items = append(items, item)
	}

	return &order, items, &patient, nil
}

// GetPatientByID retrieves patient details by ID
func (r *orderRepository) GetPatientByID(ctx context.Context, patientID uuid.UUID) (*domain.Patient, error) {
	query := `
        SELECT id, full_name, phone_number, emergency_phone_number, created_at, updated_at
        FROM patients
        WHERE id = $1
    `
	var patient domain.Patient
	err := r.db.QueryRowContext(ctx, query, patientID).Scan(
		&patient.ID, &patient.FullName, &patient.PhoneNumber, &patient.EmergencyPhoneNumber, &patient.CreatedAt, &patient.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPatientNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get patient")
		return nil, err
	}
	return &patient, nil
}

// CreateOrder creates a new order with items
func (r *orderRepository) CreateOrder(ctx context.Context, order *domain.Order, items []domain.OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	query := `
        INSERT INTO orders (id, hospital_id, patient_id, pharmacy_id, order_date, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err = tx.ExecContext(ctx, query,
		order.ID, order.HospitalID, order.PatientID, order.PharmacyID, order.OrderDate, order.Status, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to insert order")
		return err
	}

	for _, item := range items {
		itemQuery := `
            INSERT INTO order_items (id, order_id, medicine_variant_id, quantity, price_per_unit, created_at)
            VALUES ($1, $2, $3, $4, $5, $6)
        `
		_, err = tx.ExecContext(ctx, itemQuery,
			uuid.New(), order.ID, item.MedicineVariantID, item.Quantity, item.PricePerUnit, item.CreatedAt,
		)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to insert order item")
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error().Err(err).Msg("Failed to commit transaction")
		return err
	}
	return nil
}

// RequestOTP stores a new OTP for an order
func (r *orderRepository) RequestOTP(ctx context.Context, otp *domain.OrderOTP) error {
	query := `
        INSERT INTO order_otps (order_id, otp, phone_number, expires_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (order_id) DO UPDATE
        SET otp = EXCLUDED.otp, phone_number = EXCLUDED.phone_number, expires_at = EXCLUDED.expires_at, updated_at = EXCLUDED.updated_at
    `
	_, err := r.db.ExecContext(ctx, query,
		otp.OrderID, otp.OTP, otp.PhoneNumber, otp.ExpiresAt, otp.CreatedAt, otp.UpdatedAt,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to insert or update order OTP")
		return err
	}
	return nil
}

// VerifyOrder verifies the OTP for an order and updates status
func (r *orderRepository) VerifyOrder(ctx context.Context, orderID uuid.UUID, otp string) (bool, error) {
	query := `
        SELECT o.status, oo.otp, oo.expires_at
        FROM orders o
        LEFT JOIN order_otps oo ON o.id = oo.order_id
        WHERE o.id = $1
    `
	var status, storedOTP string
	var expiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(&status, &storedOTP, &expiresAt)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("order_id", orderID.String()).Msg("Order not found")
		return false, domain.ErrOrderNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get order OTP")
		return false, err
	}
	if status == "confirmed" {
		return false, domain.ErrOrderAlreadyConfirmed
	}
	if !expiresAt.Valid || time.Now().After(expiresAt.Time) || storedOTP != otp {
		return false, nil
	}
	updateQuery := `
        UPDATE orders
        SET status = 'confirmed', updated_at = $2
        WHERE id = $1 AND status = 'pending'
    `
	result, err := r.db.ExecContext(ctx, updateQuery, orderID, time.Now())
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to update order status")
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to check rows affected")
		return false, err
	}
	if rowsAffected == 0 {
		return false, nil
	}
	return true, nil
}
