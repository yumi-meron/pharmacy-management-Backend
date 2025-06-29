package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pharmacist-backend/delivery/http"
	"github.com/pharmacist-backend/usecase"
)

func SetupRoutes(r *gin.Engine, authUC usecase.AuthUsecase) {
	http.NewAuthHandler(r, authUC)
}
