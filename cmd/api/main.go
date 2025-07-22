package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"pharmacy-management-backend/config"
	"pharmacy-management-backend/delivery/route"
	"pharmacy-management-backend/infrastructure"
	"pharmacy-management-backend/infrastructure/middleware"
	"pharmacy-management-backend/repository"
	"pharmacy-management-backend/usecase"
	"pharmacy-management-backend/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database
	db, err := infrastructure.NewDatabase(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize Twilio service
	twilioService := infrastructure.NewTwilioService(cfg, logger)

	// Initialize validator
	v := utils.NewValidator()
	if err := utils.DebugValidator(v); err != nil {
		logger.Fatal().Err(err).Msg("Custom validations failed")
	}
	logger.Info().Msg("Custom validations registered successfully")

	// Initialize repositories
	authRepo := repository.NewAuthRepository(db, logger)
	pharmacyRepo := repository.NewPharmacyRepository(db, logger)
	medicineRepo := repository.NewMedicineRepository(db, logger)
	saleRepo := repository.NewSaleRepository(db, logger)
	orderRepo := repository.NewOrderRepository(db, logger)

	// Initialize use cases
	authUsecase := usecase.NewAuthUsecase(authRepo, twilioService, cfg)
	userUsecase := usecase.NewUserUsecase(authRepo)
	pharmacyUsecase := usecase.NewPharmacyUsecase(pharmacyRepo)
	medicineUsecase := usecase.NewMedicineUsecase(medicineRepo, pharmacyRepo)
	saleUsecase := usecase.NewSaleUsecase(saleRepo, medicineRepo)
	orderUsecase := usecase.NewOrderUsecase(orderRepo)

	// Initialize Gin router
	router := gin.Default()

	// Add logger middleware
	router.Use(middleware.LoggerMiddleware(logger))

	// Set up routes
	route.SetupRoutes(router, authUsecase, userUsecase, pharmacyUsecase, medicineUsecase, saleUsecase, orderUsecase, cfg, v)

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Info().Msgf("Starting server on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server shutdown failed")
	}
	logger.Info().Msg("Server shutdown complete")
}
