package middleware

import (
	"errors"
	"net/http"

	"pharmacist-backend/utils"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware restricts access to specified roles
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, errors.New("role not found in context"))
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, errors.New("insufficient permissions"))
		c.Abort()
	}
}
