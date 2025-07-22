package repository

import (
	"context"
	"database/sql"
	"pharmacy-management-backend/domain"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// OrderRepository defines the interface for order-related database operations
type OrderRepository interface {
	ListOrders(ctx context.Context, pharmacyID uuid.UUID, limit, offset int) ([]domain.OrderResponse, error)
	GetOrderDetails(ctx context.Context, orderID uuid.UUID) (*domain.Order, []domain.OrderItem, *domain.Patient, error)
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
        SELECT o.id, h.name, p.full_name, o.order_date
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
		if err := rows.Scan(&order.ID, &order.HospitalName, &order.PatientName, &order.OrderDate); err != nil {
			r.logger.Error().Err(err).Msg("Failed to scan order")
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// GetOrderDetails retrieves order details including patient and items
func (r *orderRepository) GetOrderDetails(ctx context.Context, orderID uuid.UUID) (*domain.Order, []domain.OrderItem, *domain.Patient, error) {
	// Get order
	query := `
        SELECT o.id, o.hospital_id, o.patient_id, o.pharmacy_id, o.order_date, o.created_at, o.updated_at
        FROM orders o
        WHERE o.id = $1
    `
	var order domain.Order
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.HospitalID, &order.PatientID, &order.PharmacyID, &order.OrderDate, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		r.logger.Info().Str("order_id", orderID.String()).Msg("Order not found")
		return nil, nil, nil, domain.ErrOrderNotFound
	}
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get order")
		return nil, nil, nil, err
	}

	// Get patient
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

	// Get order items with medicine details
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
