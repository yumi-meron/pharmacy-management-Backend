package domain

import (
	"time"

	"github.com/google/uuid"
)

// Hospital represents a hospital entity
type Hospital struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Patient represents a patient entity
type Patient struct {
	ID                   uuid.UUID `json:"id"`
	FullName             string    `json:"full_name"`
	PhoneNumber          string    `json:"phone_number"`
	EmergencyPhoneNumber string    `json:"emergency_phone_number"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Order represents an order from a hospital
type Order struct {
	ID         uuid.UUID `json:"id"`
	HospitalID uuid.UUID `json:"hospital_id"`
	PatientID  uuid.UUID `json:"patient_id"`
	PharmacyID uuid.UUID `json:"pharmacy_id"`
	OrderDate  time.Time `json:"order_date"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// OrderOTP represents OTP details for an order
type OrderOTP struct {
	OrderID     uuid.UUID `json:"order_id"`
	OTP         string    `json:"otp"`
	PhoneNumber string    `json:"phone_number"`
	ExpiresAt   time.Time `json:"otp_expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID                uuid.UUID `json:"id"`
	OrderID           uuid.UUID `json:"order_id"`
	MedicineVariantID uuid.UUID `json:"medicine_variant_id"`
	Quantity          int       `json:"quantity"`
	PricePerUnit      float64   `json:"price_per_unit"`
	CreatedAt         time.Time `json:"created_at"`
	MedicineName      string    `json:"medicine_name,omitempty"`
	Unit              string    `json:"unit,omitempty"`
}

// OrderResponse defines the response for listing orders
type OrderResponse struct {
	ID           uuid.UUID `json:"id"`
	HospitalName string    `json:"hospital_name"`
	PatientName  string    `json:"patient_name"`
	OrderDate    time.Time `json:"order_date"`
	Status       string    `json:"status"`
}

// PatientResponse defines the patient details in order details
type PatientResponse struct {
	ID                   uuid.UUID `json:"id"`
	FullName             string    `json:"full_name"`
	PhoneNumber          string    `json:"phone_number"`
	EmergencyPhoneNumber string    `json:"emergency_phone_number"`
}

// OrderItemResponse defines an item in the order details response
type OrderItemResponse struct {
	MedicineName string  `json:"medicine_name"`
	Unit         string  `json:"unit"`
	Quantity     int     `json:"quantity"`
	PricePerUnit float64 `json:"price_per_unit"`
}

// OrderDetailsResponse defines the response for order details
type OrderDetailsResponse struct {
	Patient    PatientResponse     `json:"patient"`
	Items      []OrderItemResponse `json:"items"`
	TotalPrice float64             `json:"total_price"`
}

// CreateOrderRequest defines the request for creating an order
type CreateOrderRequest struct {
	HospitalID uuid.UUID          `json:"hospital_id" validate:"required"`
	PatientID  uuid.UUID          `json:"patient_id" validate:"required"`
	Items      []OrderItemRequest `json:"items" validate:"required,dive"`
}

// OrderItemRequest defines an item in the create order request
type OrderItemRequest struct {
	MedicineVariantID uuid.UUID `json:"medicine_variant_id" validate:"required"`
	Quantity          int       `json:"quantity" validate:"required,gt=0"`
	PricePerUnit      float64   `json:"price_per_unit" validate:"required,gte=0"`
}

// RequestOTPRequest defines the request for sending an OTP
type RequestOTPRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,phone"`
}

// SMSService defines the interface for sending SMS
type SMSService interface {
	SendSMS(to, body string) error
}
