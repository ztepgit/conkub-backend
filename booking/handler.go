// booking/handler.go
package booking

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// BookSeat จัดการ API /bookings สำหรับจองที่นั่งและสร้าง Stripe Checkout URL
func (h *Handler) BookSeat(c *gin.Context) {
	// ดึง UserID ที่ได้จาก Middleware ตรวจสอบ Supabase JWT
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req BookSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// ตั้ง Timeout ป้องกัน API ค้างนานเกินไป (Goroutine Management)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// เรียก Service และรับ checkoutURL กลับมา
	checkoutURL, err := h.service.BookSeat(ctx, userID, req)
	if err != nil {
		// จัดการ Error Message ตามที่ Service โยนมา
		if err.Error() == "seat already booked" || err.Error() == "seat is currently being booked by someone else" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // HTTP 409 Conflict
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔴 [B] ตรวจสอบว่า Handler ส่ง URL กลับไปใน JSON Payload หรือไม่
	log.Printf("[Stripe] Returning checkout URL: %s", checkoutURL)

	// ส่ง URL กลับไปให้หน้าบ้าน เพื่อให้ Next.js ทำการ Redirect ไปจ่ายเงิน
	c.JSON(http.StatusOK, gin.H{
		"message": "seat locked, proceed to payment",
		"url":     checkoutURL,
	})
}

// StripeWebhook รับ Event จาก Stripe เมื่อมีการจ่ายเงินสำเร็จ
func (h *Handler) StripeWebhook(c *gin.Context) {
	// 1. ดึง Stripe-Signature จาก Header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Stripe-Signature header"})
		return
	}

	// 2. 🔴 อ่าน Raw Body ห้ามแปลงเป็น JSON เด็ดขาด เพราะต้องใช้ Verify Signature
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// 3. ส่งให้ Service ประมวลผล
	err = h.service.ProcessStripeWebhook(c.Request.Context(), payload, signature)
	if err != nil {
		log.Printf("[Stripe Webhook Error] %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook processing failed"})
		return
	}

	// 4. ตอบ 200 OK เสมอหากสำเร็จ เพื่อให้ Stripe ทราบว่าเราได้รับแล้ว
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
