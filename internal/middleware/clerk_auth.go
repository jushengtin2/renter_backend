package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)
func OptionalClerkAuth() gin.HandlerFunc {
	jwksURL := "https://giving-cobra-50.clerk.accounts.dev/.well-known/jwks.json"
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		panic("failed to get Clerk JWKS: " + err.Error())
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Set("clerk_user_id", nil)
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, jwks.Keyfunc)
		if err != nil || !token.Valid {
			c.Set("clerk_user_id", nil)
			c.Next()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("clerk_user_id", claims["sub"])
		}
		
		c.Next()
	}
}

func ClerkAuth() gin.HandlerFunc {
	jwksURL := "https://giving-cobra-50.clerk.accounts.dev/.well-known/jwks.json"

	// 初始化 JWKS，會自動 refresh
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		panic("failed to get Clerk JWKS: " + err.Error())
	}

	return func(c *gin.Context) {
		// 1. 讀取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. 驗證 JWT
		token, err := jwt.Parse(tokenStr, jwks.Keyfunc)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// 3. 提取 user id
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("clerk_user_id", claims["sub"])
			fmt.Println("clerk_user_id:", claims["sub"])
		}

		// 4. 放行
		c.Next()
	}
}


