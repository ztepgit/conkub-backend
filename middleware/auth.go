// middleware/auth.go
package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth เป็น Middleware ตรวจสอบ JWT จาก Supabase
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. ดึง Token จาก Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// รูปแบบต้องเป็น "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]

		// 2. Parse และ Validate JWT ด้วย Secret ของ Supabase
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			fmt.Println("🔐 JWT Algorithm:", token.Header["alg"])
			// ตรวจสอบ Signing Method ว่าเป็น HMAC ตามที่ Supabase ใช้หรือไม่
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			fmt.Println("❌ JWT Parse Error:", err)
		}

		if token != nil {
			fmt.Println("🔐 JWT Valid:", token.Valid)
		}

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 3. ดึง User ID (sub) ออกมาจาก Token Claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		userID, ok := claims["sub"].(string) // "sub" ใน Supabase คือ UUID ของ User
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			return
		}

		// 4. ส่ง UserID ไปให้ Handler ต่อไปใช้งาน (เช่น ตอนบันทึกข้อมูลการจอง)
		c.Set("userID", userID)
		c.Next()
	}
}
