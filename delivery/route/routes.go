package route

import (
	"github.com/gin-gonic/gin"
	"github.com/yumi-meron/pharmacy-management-app/pharmacist-backend/module/delivery/http"
	"github.com/yumi-meron/pharmacy-management-app/pharmacist-backend/module/usecase"
)

func SetupRoutes(r *gin.Engine, authUC usecase.AuthUsecase) {
	http.NewAuthHandler(r, authUC)
}
