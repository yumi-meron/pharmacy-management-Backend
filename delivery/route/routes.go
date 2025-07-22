package route

import (
	"pharmacy-management-backend/config"
	"pharmacy-management-backend/delivery/http"
	"pharmacy-management-backend/infrastructure/middleware"
	"pharmacy-management-backend/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SetupRoutes configures the application routes
func SetupRoutes(
	r *gin.Engine,
	authUsecase usecase.AuthUsecase,
	userUsecase usecase.UserUsecase,
	pharmacyUsecase usecase.PharmacyUsecase,
	medicineUsecase usecase.MedicineUsecase,
	saleUsecase usecase.SaleUsecase,
	orderUsecase usecase.OrderUsecase,
	cfg *config.Config,
	validator *validator.Validate,
) {
	// Initialize handlers
	authHandler := http.NewAuthHandler(authUsecase, validator)
	userHandler := http.NewUserHandler(userUsecase, validator)
	pharmacyHandler := http.NewPharmacyHandler(pharmacyUsecase, validator)
	medicineHandler := http.NewMedicineHandler(medicineUsecase, validator)
	saleHandler := http.NewSaleHandler(saleUsecase, validator)
	orderHandler := http.NewOrderHandler(orderUsecase, validator)

	// Middleware
	authMiddleware := middleware.AuthMiddleware(cfg)
	adminMiddleware := middleware.RoleMiddleware("admin")
	adminOwnerMiddleware := middleware.RoleMiddleware("admin", "owner")
	saleMiddleware := middleware.RoleMiddleware("owner", "pharmacist")

	// Auth routes (public)
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/forgot-password", authHandler.RequestPasswordReset)
		auth.POST("/reset-password", authHandler.ResetPassword)
		auth.POST("/refresh-token", authHandler.RefreshToken)
	}

	// User routes (protected)
	users := r.Group("/api/users")
	users.Use(authMiddleware)
	{
		users.GET("/me", authHandler.GetProfile)
		users.PUT("/me", authHandler.UpdateProfile)
		users.POST("/owners", adminMiddleware, userHandler.CreateOwner)
		users.POST("/pharmacists", adminOwnerMiddleware, userHandler.CreatePharmacist)
		users.GET("/pharmacists", adminOwnerMiddleware, userHandler.ListPharmacists)
	}

	// Pharmacy routes (protected)
	pharmacies := r.Group("/api/pharmacies")
	pharmacies.Use(authMiddleware)
	{
		pharmacies.POST("/", adminMiddleware, pharmacyHandler.Create)
		pharmacies.GET("/", pharmacyHandler.GetAll)
		pharmacies.GET("/:id", pharmacyHandler.GetByID)
		pharmacies.PUT("/:id", adminOwnerMiddleware, pharmacyHandler.Update)
		pharmacies.DELETE("/:id", adminMiddleware, pharmacyHandler.Delete)
	}

	// Medicine routes (protected)
	medicines := r.Group("/api/medicines")
	medicines.Use(authMiddleware)
	{
		medicines.POST("/", adminOwnerMiddleware, medicineHandler.Create)
		medicines.GET("/", medicineHandler.GetAll)
		medicines.GET("/search", saleMiddleware, saleHandler.SearchMedicines)
		medicines.GET("/:id", medicineHandler.GetByID)
		medicines.PUT("/:id", adminOwnerMiddleware, medicineHandler.Update)
		medicines.DELETE("/:id", adminMiddleware, medicineHandler.Delete)
		medicines.POST("/:id/variants", adminOwnerMiddleware, medicineHandler.CreateVariant)
		medicines.GET("/:id/variants", medicineHandler.GetVariants)
		medicines.GET("/:id/variants/:variant_id", medicineHandler.GetVariantByID)
		medicines.PUT("/:id/variants/:variant_id", adminOwnerMiddleware, medicineHandler.UpdateVariant)
		medicines.DELETE("/:id/variants/:variant_id", adminMiddleware, medicineHandler.DeleteVariant)
	}

	// Sale routes (protected)
	sales := r.Group("/api/sales")
	sales.Use(authMiddleware, saleMiddleware)
	{
		sales.POST("/", saleHandler.ConfirmSale)
		sales.GET("/", saleHandler.GetSales)
		sales.GET("/:id/receipt", saleHandler.GetReceipt)
	}

	// Cart routes (protected)
	cart := r.Group("/api/cart")
	cart.Use(authMiddleware, saleMiddleware)
	{
		cart.POST("/", saleHandler.AddToCart)
		cart.GET("/", saleHandler.GetCart)
		cart.DELETE("/:item_id", saleHandler.RemoveFromCart)
	}

	// Order routes (protected)
	orders := r.Group("/api/orders")
	orders.Use(authMiddleware, middleware.RoleMiddleware("admin", "owner", "pharmacist"))
	{
		orders.GET("", orderHandler.ListOrders)
		orders.GET("/:id", orderHandler.GetOrderDetails)
	}
}
