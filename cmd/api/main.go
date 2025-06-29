package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/yumi-meron/pharmacy-management-app/pharmacist-backend/module/delivery/route"
	"github.com/yumi-meron/pharmacy-management-app/pharmacist-backend/module/repository"
	"github.com/yumi-meron/pharmacy-management-app/pharmacist-backend/module/usecase"
)

func main() {
	_ = godotenv.Load()

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	r := gin.Default()

	// DI setup
	authRepo := repository.NewAuthRepository(db)
	authUC := usecase.NewAuthUsecase(authRepo)

	route.SetupRoutes(r, authUC)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
