package middleware

import (
	"net/http"

	"pharmacist-backend/domain"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware restricts access to routes based on user role
func RoleMiddleware(requiredRole domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": domain.ErrUnauthorized.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}
