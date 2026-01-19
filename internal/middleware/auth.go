package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{"error": "missing token"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			fmt.Println(token)
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{"error": "invalid token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		role := claims["role"].(string)

		for _, requiredRole := range requiredRoles {
			if requiredRole != "" && role != requiredRole {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{"error": "forbidden"})
				return
			}
		}

		c.Set("userID", int(claims["userID"].(float64)))
		c.Set("role", role)

		c.Next()
	}
}
