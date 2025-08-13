package middleware

import (
	"errors"
	"net/http"
	"strings"

	"pharmacy-management-backend/config"
	"pharmacy-management-backend/domain"
	"pharmacy-management-backend/usecase"
	"pharmacy-management-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(cfg *config.Config, usecase usecase.AuthUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, errors.New("authorization header required"))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, errors.New("invalid authorization header format"))
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// Check if token is blacklisted via usecase
		blacklisted, err := usecase.IsTokenBlacklisted(c.Request.Context(), tokenStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
			c.Abort()
			return
		}
		if blacklisted {
			utils.ErrorResponse(c, http.StatusUnauthorized, errors.New("token has been revoked"))
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, domain.ErrUnauthorized
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			utils.ErrorResponse(c, http.StatusUnauthorized, errors.New("invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, errors.New("invalid token claims"))
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])
		c.Set("pharmacy_id", claims["pharmacy_id"])
		c.Next()
	}
}
