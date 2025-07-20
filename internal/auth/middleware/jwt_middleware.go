package middleware

import (
	"github/english-app/internal/auth/token"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Pass jwtMaker so you can use the secret inside middleware
func AuthMiddleware(jwtMaker *token.JWTMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtMaker.VerifyToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Set user info in context for use in protected handlers
		c.Set("user_id", claims.UserId)

		c.Next()
	}
}
