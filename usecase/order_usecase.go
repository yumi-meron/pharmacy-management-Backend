package usecase

import (
	"context"
	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/repository"
	"pharmacy-management-backend/utils"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// OrderUsecase defines the interface for order-related business logic
type OrderUsecase interface {
	ListOrders(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, limit, offset int) ([]domain.OrderResponse, error)
	GetOrderDetails(ctx context.Context, callerRole string, callerPharmacyID, orderID uuid.UUID) (*domain.OrderDetailsResponse, error)
	CreateOrder(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, req domain.CreateOrderRequest) (*domain.Order, error)
	RequestOTP(ctx context.Context, callerRole string, callerPharmacyID, orderID uuid.UUID, req domain.RequestOTPRequest) error
	VerifyOrder(ctx context.Context, callerRole string, callerPharmacyID, orderID uuid.UUID, otp string) error
}

// orderUsecase implements OrderUsecase
type orderUsecase struct {
	repo       repository.OrderRepository
	smsService domain.SMSService
	logger     zerolog.Logger
}

// NewOrderUsecase creates a new OrderUsecase
func NewOrderUsecase(repo repository.OrderRepository, smsService domain.SMSService, logger zerolog.Logger) OrderUsecase {
	return &orderUsecase{repo, smsService, logger}
}

// ListOrders retrieves a list of orders
func (u *orderUsecase) ListOrders(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, limit, offset int) ([]domain.OrderResponse, error) {
	if !isValidRole(callerRole) {
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
	if !isValidRole(callerRole) {
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

// CreateOrder creates a new order
func (u *orderUsecase) CreateOrder(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, req domain.CreateOrderRequest) (*domain.Order, error) {
	if !isValidRole(callerRole) {
		return nil, domain.ErrUnauthorized
	}

	_, err := u.repo.GetPatientByID(ctx, req.PatientID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	order := &domain.Order{
		ID:         uuid.New(),
		HospitalID: req.HospitalID,
		PatientID:  req.PatientID,
		PharmacyID: callerPharmacyID,
		OrderDate:  now,
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	var items []domain.OrderItem
	for _, itemReq := range req.Items {
		items = append(items, domain.OrderItem{
			ID:                uuid.New(),
			OrderID:           order.ID,
			MedicineVariantID: itemReq.MedicineVariantID,
			Quantity:          itemReq.Quantity,
			PricePerUnit:      itemReq.PricePerUnit,
			CreatedAt:         now,
		})
	}

	if err := u.repo.CreateOrder(ctx, order, items); err != nil {
		u.logger.Error().Err(err).Msg("Failed to create order")
		return nil, err
	}

	return order, nil
}

// RequestOTP generates and sends an OTP to the specified phone number
func (u *orderUsecase) RequestOTP(ctx context.Context, callerRole string, callerPharmacyID, orderID uuid.UUID, req domain.RequestOTPRequest) error {
	if !isValidRole(callerRole) {
		return domain.ErrUnauthorized
	}

	order, _, patient, err := u.repo.GetOrderDetails(ctx, orderID)
	if err != nil {
		return err
	}

	if callerRole != string(domain.RoleAdmin) && order.PharmacyID != callerPharmacyID {
		return domain.ErrUnauthorized
	}

	if req.PhoneNumber != patient.PhoneNumber && req.PhoneNumber != patient.EmergencyPhoneNumber {
		return domain.ErrInvalidPhone
	}

	now := time.Now()
	otp := &domain.OrderOTP{
		OrderID:     orderID,
		OTP:         utils.GenerateOTP(),
		PhoneNumber: req.PhoneNumber,
		ExpiresAt:   now.Add(5 * time.Minute),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.repo.RequestOTP(ctx, otp); err != nil {
		u.logger.Error().Err(err).Msg("Failed to store OTP")
		return err
	}

	err = u.smsService.SendSMS(req.PhoneNumber, "Your order OTP is: "+otp.OTP)
	if err != nil {
		u.logger.Error().Err(err).Msg("Failed to send OTP SMS")
		return err
	}

	return nil
}

// VerifyOrder verifies the OTP for an order
func (u *orderUsecase) VerifyOrder(ctx context.Context, callerRole string, callerPharmacyID, orderID uuid.UUID, otp string) error {
	if !isValidRole(callerRole) {
		return domain.ErrUnauthorized
	}

	order, _, _, err := u.repo.GetOrderDetails(ctx, orderID)
	if err != nil {
		return err
	}

	if callerRole != string(domain.RoleAdmin) && order.PharmacyID != callerPharmacyID {
		return domain.ErrUnauthorized
	}

	success, err := u.repo.VerifyOrder(ctx, orderID, otp)
	if err != nil {
		return err
	}
	if !success {
		return domain.ErrInvalidOTP
	}
	return nil
}

// isValidRole checks if the role is valid for order operations
func isValidRole(role string) bool {
	return role == string(domain.RoleAdmin) || role == string(domain.RoleOwner) || role == string(domain.RolePharmacist)
}
