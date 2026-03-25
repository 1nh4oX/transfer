package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"transfer/backend/internal/model"
)

const CurrentUserKey = "currentUser"

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Invalid token format."})
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if sub, ok := claims["sub"].(string); ok && strings.TrimSpace(sub) != "" {
				c.Set(CurrentUserKey, sub)
			}
		}

		c.Next()
	}
}

func CurrentUser(c *gin.Context) (string, bool) {
	v, ok := c.Get(CurrentUserKey)
	if !ok {
		return "", false
	}

	username, ok := v.(string)
	if !ok || strings.TrimSpace(username) == "" {
		return "", false
	}

	return username, true
}
