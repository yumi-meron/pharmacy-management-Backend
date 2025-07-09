package route

import (
	"pharmacist-backend/config"
	"pharmacist-backend/delivery/http"
	"pharmacist-backend/infrastructure/middleware"
	"pharmacist-backend/usecase"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SetupRoutes configures the application routes
func SetupRoutes(
	r *gin.Engine,
	authUsecase usecase.AuthUsecase,
	userUsecase usecase.UserUsecase,
	pharmacyUsecase usecase.PharmacyUsecase,
	cfg *config.Config,
	validator *validator.Validate,
) {
	// Initialize handlers
	authHandler := http.NewAuthHandler(authUsecase, validator)
	userHandler := http.NewUserHandler(userUsecase, validator)
	pharmacyHandler := http.NewPharmacyHandler(pharmacyUsecase, validator)

	// Middleware
	authMiddleware := middleware.AuthMiddleware(cfg)
	adminMiddleware := middleware.RoleMiddleware("admin")
	adminOwnerMiddleware := middleware.RoleMiddleware("admin", "owner")

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
}
