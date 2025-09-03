package usecase

import (
	"context"
	"errors"
	"strings"

	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/repository"
)

// BarcodeUsecase defines the interface for barcode-related business logic
type BarcodeUsecase interface {
	LookupBarcode(ctx context.Context, barcodeValue string) (*domain.Medicine, error)
	AddBarcodes(ctx context.Context, barcodes domain.Barcode) error
}

// barcodeUsecase implements BarcodeUsecase
type barcodeUsecase struct {
	repo    repository.BarcodeRepository
	medrepo repository.MedicineRepository
}

// NewBarcodeUsecase creates a new BarcodeUsecase
func NewBarcodeUsecase(repo repository.BarcodeRepository, medrepo repository.MedicineRepository) BarcodeUsecase {
	return &barcodeUsecase{repo: repo, medrepo: medrepo}
}

// LookupBarcode looks up a barcode using an external API and returns product details
func (u *barcodeUsecase) LookupBarcode(ctx context.Context, barcodeValue string) (*domain.Medicine, error) {
	if strings.TrimSpace(barcodeValue) == "" {
		return nil, errors.New("barcode value cannot be empty")
	}
	// Call the repository to lookup the barcode
	medicine_id, err := u.repo.LookupBarcode(ctx, barcodeValue)
	if err != nil {
		return nil, err
	}
	medicine, err := u.medrepo.GetByID(ctx, medicine_id)
	if err != nil {
		return nil, err
	}
	return medicine, nil
}

// AddBarcodes adds multiple barcodes to the database
func (u *barcodeUsecase) AddBarcodes(ctx context.Context, barcodes domain.Barcode) error {

	// Call the repository to add the barcodes
	if err := u.repo.AddBarcodes(ctx, barcodes); err != nil {
		return err
	}
	return nil
}
