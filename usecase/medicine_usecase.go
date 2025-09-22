package usecase

import (
	"context"
	"strings"
	"time"

	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/repository"

	"github.com/google/uuid"
)

// MedicineUsecase defines the interface for medicine-related business logic
type MedicineUsecase interface {
	Create(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, input domain.CreateMedicineInput) error
	GetAll(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID) ([]domain.Medicine, error)
	GetByID(ctx context.Context, callerRole string, callerPharmacyID, id uuid.UUID) (*domain.Medicine, error)
	Delete(ctx context.Context, callerRole string, id uuid.UUID) error
	CreateVariant(ctx context.Context, callerRole string, callerPharmacyID, medicineID uuid.UUID, input domain.CreateMedicineVariantInput) error
	GetVariants(ctx context.Context, callerRole string, callerPharmacyID, medicineID uuid.UUID) ([]domain.MedicineVariant, error)
	GetVariantByID(ctx context.Context, callerRole string, callerPharmacyID, medicineID, variantID uuid.UUID) (*domain.MedicineVariant, error)
	UpdateVariant(ctx context.Context, callerRole string, callerPharmacyID, medicineID, variantID uuid.UUID, input domain.UpdateMedicineVariantInput) (*domain.Medicine, error)
	DeleteVariant(ctx context.Context, callerRole string, variantID uuid.UUID) error
	SearchMedicines(ctx context.Context, params domain.SearchParams) ([]domain.Medicine, error)
	UploadPicture(ctx context.Context, userID uuid.UUID, fileData []byte, fileExt string) (string, error)
}

// medicineUsecase implements MedicineUsecase
type medicineUsecase struct {
	repo         repository.MedicineRepository
	pharmacyRepo repository.PharmacyRepository
}

// NewMedicineUsecase creates a new MedicineUsecase
func NewMedicineUsecase(repo repository.MedicineRepository, pharmacyRepo repository.PharmacyRepository) MedicineUsecase {
	return &medicineUsecase{repo, pharmacyRepo}
}

// Create creates a new medicine
func (u *medicineUsecase) Create(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID, input domain.CreateMedicineInput) error {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) {
		return domain.ErrUnauthorized
	}

	// Verify pharmacy exists
	if _, err := u.pharmacyRepo.GetByID(ctx, input.PharmacyID); err != nil {
		return domain.ErrInvalidPharmacy
	}

	// Restrict Owners to their own pharmacy
	if callerRole == string(domain.RoleOwner) && callerPharmacyID != input.PharmacyID {
		return domain.ErrUnauthorized
	}

	medicine := domain.Medicine{
		ID:          uuid.New(),
		PharmacyID:  input.PharmacyID,
		Name:        input.Name,
		Description: input.Description,
		Picture:     input.Picture,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	return u.repo.Create(ctx, medicine)
}

// GetAll retrieves medicines based on role
func (u *medicineUsecase) GetAll(ctx context.Context, callerRole string, callerPharmacyID uuid.UUID) ([]domain.Medicine, error) {
	pharmacyID := callerPharmacyID
	// if callerRole != string(domain.RoleAdmin) {
	// }
	return u.repo.GetAll(ctx, pharmacyID)
}

// GetByID retrieves a medicine with role-based restrictions
func (u *medicineUsecase) GetByID(ctx context.Context, callerRole string, callerPharmacyID, id uuid.UUID) (*domain.Medicine, error) {
	medicine, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if callerRole != string(domain.RoleAdmin) && callerPharmacyID != medicine.PharmacyID {
		return nil, domain.ErrUnauthorized
	}
	return medicine, nil
}

// // Update updates a medicine with role-based restrictions
// func (u *medicineUsecase) Update(ctx context.Context, callerRole string, callerPharmacyID, id uuid.UUID, input domain.UpdateMedicineInput) error {
// 	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) {
// 		return domain.ErrUnauthorized
// 	}

// 	medicine, err := u.repo.GetByID(ctx, id)
// 	if err != nil {
// 		return err
// 	}

// 	if callerRole == string(domain.RoleOwner) && callerPharmacyID != medicine.PharmacyID {
// 		return domain.ErrUnauthorized
// 	}

// 	medicine.Name = input.Name
// 	medicine.Description = input.Description
// 	medicine.Picture = input.Picture
// 	medicine.UpdatedAt = time.Now().UTC()

// 	return u.repo.Update(ctx, *medicine)
// }

// Delete deletes a medicine (Admin-only)
func (u *medicineUsecase) Delete(ctx context.Context, callerRole string, id uuid.UUID) error {
	if callerRole != string(domain.RoleAdmin) {
		return domain.ErrUnauthorized
	}

	count, err := u.repo.CountVariants(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrMedicineHasVariants
	}

	return u.repo.Delete(ctx, id)
}

// CreateVariant creates a new medicine variant
func (u *medicineUsecase) CreateVariant(ctx context.Context, callerRole string, callerPharmacyID, medicineID uuid.UUID, input domain.CreateMedicineVariantInput) error {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) {
		return domain.ErrUnauthorized
	}

	medicine, err := u.repo.GetByID(ctx, medicineID)
	if err != nil {
		return err
	}

	if callerRole == string(domain.RoleOwner) && callerPharmacyID != medicine.PharmacyID {
		return domain.ErrUnauthorized
	}

	if exists, err := u.repo.CheckBarcodeExists(ctx, input.Barcode); err != nil {
		return err
	} else if exists {
		return domain.ErrBarcodeTaken
	}

	variant := domain.MedicineVariant{
		ID:           uuid.New(),
		MedicineID:   medicineID,
		Brand:        input.Brand,
		Barcode:      input.Barcode,
		Unit:         input.Unit,
		PricePerUnit: input.PricePerUnit,
		ExpiryDate:   input.ExpiryDate,
		Stock:        input.Stock,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	return u.repo.CreateVariant(ctx, variant)
}

// GetVariants retrieves variants for a medicine
func (u *medicineUsecase) GetVariants(ctx context.Context, callerRole string, callerPharmacyID, medicineID uuid.UUID) ([]domain.MedicineVariant, error) {
	medicine, err := u.repo.GetByID(ctx, medicineID)
	if err != nil {
		return nil, err
	}

	if callerRole != string(domain.RoleAdmin) && callerPharmacyID != medicine.PharmacyID {
		return nil, domain.ErrUnauthorized
	}

	return u.repo.GetVariantsByMedicineID(ctx, medicineID)
}

// GetVariantByID retrieves a variant with role-based restrictions
func (u *medicineUsecase) GetVariantByID(ctx context.Context, callerRole string, callerPharmacyID, medicineID, variantID uuid.UUID) (*domain.MedicineVariant, error) {
	variant, err := u.repo.GetVariantByID(ctx, variantID)
	if err != nil {
		return nil, err
	}

	medicine, err := u.repo.GetByID(ctx, medicineID)
	if err != nil {
		return nil, err
	}

	if variant.MedicineID != medicineID {
		return nil, domain.ErrVariantNotFound
	}

	if callerRole != string(domain.RoleAdmin) && callerPharmacyID != medicine.PharmacyID {
		return nil, domain.ErrUnauthorized
	}

	return variant, nil
}

// UpdateVariant updates a medicine variant
func (u *medicineUsecase) UpdateVariant(ctx context.Context, callerRole string, callerPharmacyID, medicineID, variantID uuid.UUID, input domain.UpdateMedicineVariantInput) (*domain.Medicine, error) {
	if callerRole != string(domain.RoleAdmin) && callerRole != string(domain.RoleOwner) {
		return nil, domain.ErrUnauthorized
	}

	medicine, err := u.repo.GetByID(ctx, medicineID)
	if err != nil {
		return nil, err
	}

	if callerRole == string(domain.RoleOwner) && callerPharmacyID != medicine.PharmacyID {
		return nil, domain.ErrUnauthorized
	}

	variant, err := u.repo.GetVariantByID(ctx, variantID)
	if err != nil {
		return nil, err
	}

	if variant.MedicineID != medicineID {
		return nil, domain.ErrVariantNotFound
	}

	if input.Barcode != "" && input.Barcode != variant.Barcode {
		if exists, err := u.repo.CheckBarcodeExists(ctx, input.Barcode); err != nil {
			return nil, err
		} else if exists {
			return nil, domain.ErrBarcodeTaken
		}
	}
	//update medicine fields

	if input.Description != "" {
		medicine.Description = input.Description
	}

	if input.MedicalUsage != "" {
		medicine.MedicalUsage = input.MedicalUsage
	}

	if input.Name != "" {
		medicine.Name = input.Name
	}

	if input.Picture != "" {
		medicine.Picture = input.Picture
	}

	if err := u.repo.Update(ctx, *medicine); err != nil {
		return nil, err
	}

	if input.Brand != "" {
		variant.Brand = input.Brand
	}
	if input.Barcode != "" {
		variant.Barcode = input.Barcode
	}
	if input.Unit != "" {
		variant.Unit = input.Unit
	}
	if input.PricePerUnit > 0 {
		variant.PricePerUnit = input.PricePerUnit
	}
	if !input.ExpiryDate.IsZero() {
		variant.ExpiryDate = input.ExpiryDate
	}
	if input.Stock >= 0 {
		variant.Stock = input.Stock
	}
	variant.UpdatedAt = time.Now().UTC()

	err = u.repo.UpdateVariant(ctx, *variant)
	if err != nil {
		return nil, err
	}
	medicine, err = u.repo.GetByID(ctx, medicineID)
	if err != nil {
		return nil, err
	}

	// Return the modified medicine
	return medicine, nil
}

func (u *medicineUsecase) UploadPicture(ctx context.Context, userID uuid.UUID, fileData []byte, fileExt string) (string, error) {
	return u.repo.UploadPicture(ctx, userID, fileData, fileExt)
}

// DeleteVariant deletes a medicine variant (Admin-only)
func (u *medicineUsecase) DeleteVariant(ctx context.Context, callerRole string, variantID uuid.UUID) error {
	if callerRole != string(domain.RoleAdmin) {
		return domain.ErrUnauthorized
	}

	return u.repo.DeleteVariant(ctx, variantID)
}

func (u *medicineUsecase) SearchMedicines(ctx context.Context, params domain.SearchParams) ([]domain.Medicine, error) {

	if params.Filter != "brand" && params.Filter != "description" {
		params.Filter = "name"
	}

	if params.Catagory == "Antibiotic" || params.Catagory == "Vitamins" || params.Catagory == "Pain Relief" || params.Catagory == "Cold & Flu" {

	} else {
		params.Catagory = "None"
	}

	// Set default pagination
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// Preprocess query for prefix matching (add :* to each term)
	if params.Query != "" {
		terms := strings.Fields(params.Query)
		for i, term := range terms {
			terms[i] = term + ":*"
		}
		params.Query = strings.Join(terms, " & ")
	}
	return u.repo.SearchMedicines(ctx, params)
}
