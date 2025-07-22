package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/repository"

	"github.com/google/uuid"
)

// SaleUsecase defines the interface for sale-related business logic
type SaleUsecase interface {
	SearchMedicines(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, query string) ([]domain.MedicineVariant, error)
	AddToCart(ctx context.Context, callerRole string, callerUserID, callerPharmacyID uuid.UUID, input domain.CreateCartInput) error
	RemoveFromCart(ctx context.Context, callerRole string, callerUserID, callerPharmacyID, cartID uuid.UUID) error
	ConfirmSale(ctx context.Context, callerRole string, callerUserID, callerPharmacyID uuid.UUID) (*domain.Sale, error)
	GetSales(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, limit, offset int) ([]domain.SaleResponse, error)
	GetReceipt(ctx context.Context, callerRole string, callerPharmacyID, saleID uuid.UUID) (*domain.Receipt, error)
	GetCart(ctx context.Context, callerRole string, callerUserID uuid.UUID, callerPharmacyID uuid.UUID) ([]domain.CartResponse, error)
}

// saleUsecase implements SaleUsecase
type saleUsecase struct {
	saleRepo     repository.SaleRepository
	medicineRepo repository.MedicineRepository
}

// NewSaleUsecase creates a new SaleUsecase
func NewSaleUsecase(saleRepo repository.SaleRepository, medicineRepo repository.MedicineRepository) SaleUsecase {
	return &saleUsecase{saleRepo, medicineRepo}
}

// SearchMedicines searches for medicines by name or barcode
func (u *saleUsecase) SearchMedicines(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, query string) ([]domain.MedicineVariant, error) {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) && callerRole != string(domain.RolePharmacist) {
		return nil, domain.ErrUnauthorized
	}
	return u.saleRepo.SearchMedicines(ctx, callerPharmacyID, query)
}

// AddToCart adds an item to the cart
func (u *saleUsecase) AddToCart(ctx context.Context, callerRole string, callerUserID, callerPharmacyID uuid.UUID, input domain.CreateCartInput) error {
	if callerRole != string(domain.RoleOwner) && callerRole != string(domain.RolePharmacist) {
		return domain.ErrUnauthorized
	}

	variant, err := u.medicineRepo.GetVariantByID(ctx, input.MedicineVariantID)
	if err != nil {
		return err
	}

	medicine, err := u.medicineRepo.GetByID(ctx, variant.MedicineID)
	if err != nil {
		return err
	}

	if callerRole == string(domain.RoleAdmin) || callerPharmacyID != medicine.PharmacyID {
		return domain.ErrUnauthorized
	}

	if input.Quantity > variant.Stock {
		return domain.ErrInsufficientStock
	}

	cart := domain.Cart{
		ID:                uuid.New(),
		UserID:            callerUserID,
		PharmacyID:        callerPharmacyID,
		MedicineVariantID: input.MedicineVariantID,
		Quantity:          input.Quantity,
		CreatedAt:         time.Now(),
	}
	fmt.Println("HEy")

	return u.saleRepo.AddToCart(ctx, cart)
}

// GetCart retrieves the user's cart
func (u *saleUsecase) GetCart(ctx context.Context, callerRole string, callerUserID uuid.UUID, callerPharmacyID uuid.UUID) ([]domain.CartResponse, error) {
	if callerRole != string(domain.RoleOwner) && callerRole != string(domain.RolePharmacist) {
		return nil, domain.ErrUnauthorized
	}
	carts, err := u.saleRepo.GetCart(ctx, callerUserID)
	if err != nil {
		return nil, err
	}

	var response []domain.CartResponse
	for _, cart := range carts {
		if cart.PharmacyID != callerPharmacyID {
			return nil, domain.ErrUnauthorized
		}

		variant, err := u.medicineRepo.GetVariantByID(ctx, cart.MedicineVariantID)
		if err != nil {
			return nil, err
		}

		medicine, err := u.medicineRepo.GetByID(ctx, variant.MedicineID)
		if err != nil {
			return nil, err
		}

		response = append(response, domain.CartResponse{
			ID:           cart.ID,
			Medicine:     medicine.Name,
			PricePerUnit: variant.PricePerUnit,
			Unit:         variant.Unit,
			ImageURL:     medicine.Picture,
			Quantity:     cart.Quantity,
			CreatedAt:    cart.CreatedAt,
		})
	}
	return response, nil
}

// RemoveFromCart removes an item from the cart
func (u *saleUsecase) RemoveFromCart(ctx context.Context, callerRole string, callerUserID, callerPharmacyID, cartID uuid.UUID) error {
	if callerRole != string(domain.RoleOwner) && callerRole != string(domain.RolePharmacist) {
		return domain.ErrUnauthorized
	}

	cartItems, err := u.saleRepo.GetCart(ctx, callerUserID)
	if err != nil {
		return err
	}

	for _, item := range cartItems {
		if item.ID == cartID && item.UserID == callerUserID && item.PharmacyID == callerPharmacyID {
			return u.saleRepo.RemoveFromCart(ctx, cartID)
		}
	}
	return domain.ErrCartItemNotFound
}

// ConfirmSale confirms the sale and generates a receipt
func (u *saleUsecase) ConfirmSale(ctx context.Context, callerRole string, callerUserID, callerPharmacyID uuid.UUID) (*domain.Sale, error) {
	if callerRole != string(domain.RoleOwner) && callerRole != string(domain.RolePharmacist) {
		return nil, domain.ErrUnauthorized
	}

	cartItems, err := u.saleRepo.GetCart(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return nil, errors.New("cart is empty")
	}

	var saleItems []domain.SaleItem
	var totalPrice float64
	var receiptItems []domain.ReceiptItem

	for _, cartItem := range cartItems {
		if cartItem.PharmacyID != callerPharmacyID {
			return nil, domain.ErrUnauthorized
		}

		variant, err := u.medicineRepo.GetVariantByID(ctx, cartItem.MedicineVariantID)
		if err != nil {
			return nil, err
		}

		if cartItem.Quantity > variant.Stock {
			return nil, domain.ErrInsufficientStock
		}

		medicine, err := u.medicineRepo.GetByID(ctx, variant.MedicineID)
		if err != nil {
			return nil, err
		}

		receiptItem := domain.ReceiptItem{
			Brand:        variant.Brand,
			MedicineName: medicine.Name,
			PricePerUnit: variant.PricePerUnit,
			Quantity:     cartItem.Quantity,
			Subtotal:     float64(cartItem.Quantity) * variant.PricePerUnit,
		}
		receiptItems = append(receiptItems, receiptItem)

		saleItem := domain.SaleItem{
			ID:                uuid.New(),
			SaleID:            uuid.New(), // Will be updated after sale creation
			MedicineVariantID: cartItem.MedicineVariantID,
			Quantity:          cartItem.Quantity,
			PricePerUnit:      variant.PricePerUnit,
			CreatedAt:         time.Now(),
		}
		saleItems = append(saleItems, saleItem)
		totalPrice += receiptItem.Subtotal
	}

	sale := domain.Sale{
		ID:         uuid.New(),
		UserID:     callerUserID,
		PharmacyID: callerPharmacyID,
		TotalPrice: totalPrice,
		SaleDate:   time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	receiptContent := domain.ReceiptContent{
		Items:      receiptItems,
		PharmacyID: callerPharmacyID,
		SaleDate:   sale.SaleDate,
		TotalPrice: totalPrice,
	}

	receipt := domain.Receipt{
		ID:        uuid.New(),
		SaleID:    sale.ID,
		Content:   receiptContent,
		CreatedAt: time.Now(),
	}

	// Update SaleID in sale items
	for i := range saleItems {
		saleItems[i].SaleID = sale.ID
	}

	if err := u.saleRepo.CreateSale(ctx, sale, saleItems, receipt); err != nil {
		return nil, err
	}

	if err := u.saleRepo.ClearCart(ctx, callerUserID); err != nil {
		return nil, err
	}

	return &sale, nil
}

// GetSales retrieves sales with pagination
func (u *saleUsecase) GetSales(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, limit, offset int) ([]domain.SaleResponse, error) {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) && callerRole != string(domain.RolePharmacist) {
		return nil, domain.ErrUnauthorized
	}
	var pharmacyID uuid.UUID
	if callerRole != string(domain.RoleAdmin) {
		pharmacyID = callerPharmacyID
	}
	saleItems, err := u.saleRepo.GetSales(ctx, pharmacyID, limit, offset)
	if err != nil {
		return nil, err
	}

	var response []domain.SaleResponse
	for _, item := range saleItems {
		variant, err := u.medicineRepo.GetVariantByID(ctx, item.MedicineVariantID)
		if err != nil {
			return nil, err
		}

		medicine, err := u.medicineRepo.GetByID(ctx, variant.MedicineID)
		if err != nil {
			return nil, err
		}

		response = append(response, domain.SaleResponse{
			ID:           item.ID,
			Medicine:     medicine.Name,
			PricePerUnit: item.PricePerUnit,
			Unit:         variant.Unit,
			ImageURL:     medicine.Picture,
			Quantity:     item.Quantity,
			CreatedAt:    item.CreatedAt,
		})
	}
	return response, nil
}

// GetReceipt retrieves a receipt by sale ID
func (u *saleUsecase) GetReceipt(ctx context.Context, callerRole string, callerPharmacyID, saleID uuid.UUID) (*domain.Receipt, error) {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) && callerRole != string(domain.RolePharmacist) {
		return nil, domain.ErrUnauthorized
	}

	sale, err := u.saleRepo.GetSaleByID(ctx, saleID)
	if err != nil {
		return nil, err
	}

	if callerRole != string(domain.RoleAdmin) && callerPharmacyID != sale.PharmacyID {
		return nil, domain.ErrUnauthorized
	}

	return u.saleRepo.GetReceiptBySaleID(ctx, saleID)
}
