package middleware

import (
	"net/http"
	"strings"
	"time"
	"wongnok/internal/httputil"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	authPrefix = "Bearer "
)

var jwtSecret = []byte("this-is-very-stronge-secret")

func GenerateToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func JWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, authPrefix) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "missing token"})
			return
		}

		tokenRaw := strings.TrimPrefix(authHeader, authPrefix)
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenRaw, claims, func(t *jwt.Token) (any, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Message: "invalid token"})
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Next()
	}
}
