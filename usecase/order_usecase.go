package usecase

import (
	"context"
	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/repository"

	"github.com/google/uuid"
)

// OrderUsecase defines the interface for order-related business logic
type OrderUsecase interface {
	ListOrders(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, limit, offset int) ([]domain.OrderResponse, error)
	GetOrderDetails(ctx context.Context, callerRole string, callerPharmacyID, orderID uuid.UUID) (*domain.OrderDetailsResponse, error)
}

// orderUsecase implements OrderUsecase
type orderUsecase struct {
	repo repository.OrderRepository
}

// NewOrderUsecase creates a new OrderUsecase
func NewOrderUsecase(repo repository.OrderRepository) OrderUsecase {
	return &orderUsecase{repo}
}

// ListOrders retrieves a list of orders
func (u *orderUsecase) ListOrders(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, limit, offset int) ([]domain.OrderResponse, error) {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) && callerRole != string(domain.RolePharmacist) {
		return nil, domain.ErrUnauthorized
	}
	var pharmacyID uuid.UUID
	if callerRole != string(domain.RoleAdmin) {
		pharmacyID = callerPharmacyID
	}
	return u.repo.ListOrders(ctx, pharmacyID, limit, offset)
}

// GetOrderDetails retrieves details for a specific order
func (u *orderUsecase) GetOrderDetails(ctx context.Context, callerRole string, callerPharmacyID, orderID uuid.UUID) (*domain.OrderDetailsResponse, error) {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) && callerRole != string(domain.RolePharmacist) {
		return nil, domain.ErrUnauthorized
	}

	order, items, patient, err := u.repo.GetOrderDetails(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if callerRole != string(domain.RoleAdmin) && order.PharmacyID != callerPharmacyID {
		return nil, domain.ErrUnauthorized
	}

	var response domain.OrderDetailsResponse
	response.Patient = domain.PatientResponse{
		ID:                   patient.ID,
		FullName:             patient.FullName,
		PhoneNumber:          patient.PhoneNumber,
		EmergencyPhoneNumber: patient.EmergencyPhoneNumber,
	}

	response.Items = make([]domain.OrderItemResponse, len(items))
	var totalPrice float64
	for i, item := range items {
		response.Items[i] = domain.OrderItemResponse{
			MedicineName: item.MedicineName,
			Unit:         item.Unit,
			Quantity:     item.Quantity,
			PricePerUnit: item.PricePerUnit,
		}
		totalPrice += float64(item.Quantity) * item.PricePerUnit
	}
	response.TotalPrice = totalPrice

	return &response, nil
}
