package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth รับ jwksURL เข้ามาเพื่อดึง Public Key แบบอัตโนมัติ สำหรับตรวจสอบ Token ที่เป็น ES256
func RequireAuth(jwksURL string) gin.HandlerFunc {
	// 1. สร้างตัวจัดการ JWKS และทำ Caching ไว้ในหน่วยความจำ (ลดภาระการยิง API ไปยัง Supabase ทุก Request)
	options := keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshTimeout:  time.Second * 10,
	}

	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		log.Fatalf("Failed to initialize JWKS from URL %s: %v", jwksURL, err)
	}

	return func(c *gin.Context) {
		// 2. ดึง Token จาก Header
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

		// 3. Parse และ Validate JWT ดึง Public Key จาก jwks
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// 🔴 ตรวจสอบว่าอัลกอริทึมต้องเป็น ECDSA และระบุอย่างเจาะจงว่าเป็น "ES256" เท่านั้น
			if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			if t.Method.Alg() != "ES256" {
				return nil, fmt.Errorf("expected ES256 algorithm, got: %v", t.Method.Alg())
			}

			// คืนค่า Public Key ที่ดึงมาจาก JWKS ตาม Key ID (kid) ของ Token
			return jwks.Keyfunc(t)
		})

		// กรณี Token ไม่ถูกต้อง / หมดอายุ (Parse จะตรวจ exp ให้อัตโนมัติ)
		if err != nil || !token.Valid {
			log.Printf("[Auth Middleware] Verification failed: %v", err) // Debug log ปลอดภัย ไม่ log Token
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 4. ดึง User ID (sub) ออกมาจาก Token Claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("[Auth Middleware] Failed to parse claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		userID, ok := claims["sub"].(string) // "sub" ใน Supabase คือ UUID ของ User
		if !ok || userID == "" {
			log.Printf("[Auth Middleware] 'sub' claim not found in token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			return
		}

		// 🔴 Debug Log ที่ปลอดภัย (Log การตรวจสอบอัลกอริทึมสำเร็จ และแสดง UserID)
		log.Printf("[Auth Middleware] Successfully verified ES256 token for user: %s", userID)

		// 5. ส่ง UserID ไปให้ Handler ต่อไปใช้งาน (เช่น ตอนบันทึกข้อมูลการจอง)
		c.Set("userID", userID)
		c.Next()
	}
}