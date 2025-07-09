package route

import (
	"pharmacist-backend/config"
	"pharmacist-backend/delivery/http"
	"pharmacist-backend/infrastructure/middleware"
	"pharmacist-backend/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SetupRoutes configures the API routes
func SetupRoutes(
	router *gin.Engine,
	authUsecase usecase.AuthUsecase,
	pharmacyUsecase usecase.PharmacyUsecase,
	adminUsecase usecase.AdminUsecase,
	cfg *config.Config,
	validator *validator.Validate,
) {
	// Initialize handlers
	authHandler := http.NewAuthHandler(authUsecase, validator)
	pharmacyHandler := http.NewPharmacyHandler(pharmacyUsecase, validator)
	adminHandler := http.NewAdminHandler(adminUsecase, validator)

	// API group
	api := router.Group("/api")

	// Auth routes (public)
	api.POST("/signup", authHandler.Signup)
	api.POST("/login", authHandler.Login)
	api.POST("/password/reset", authHandler.RequestPasswordReset)
	api.POST("/password/reset/confirm", authHandler.ResetPassword)
	api.POST("/token/refresh", authHandler.RefreshToken)

	// Pharmacy routes (protected)
	pharmacy := api.Group("/pharmacies")
	pharmacy.Use(middleware.AuthMiddleware(cfg))
	{
		pharmacy.POST("", pharmacyHandler.Create)
		pharmacy.GET("", pharmacyHandler.GetAll)
		pharmacy.GET("/:id", pharmacyHandler.GetByID)
		pharmacy.PUT("/:id", pharmacyHandler.Update)
		pharmacy.DELETE("/:id", middleware.RoleMiddleware("admin"), pharmacyHandler.Delete)
	}

	// Admin routes (protected, admin only)
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("admin"))
	{
		admin.POST("/users", adminHandler.CreateUser)
	}
}
