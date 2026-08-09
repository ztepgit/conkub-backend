// booking/handler.go
package booking

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
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

	// ส่ง URL กลับไปให้หน้าบ้าน เพื่อให้ Next.js ทำการ Redirect ไปจ่ายเงิน
	c.JSON(http.StatusOK, gin.H{
		"message": "seat locked, proceed to payment",
		"url":     checkoutURL,
	})
}

// StripeWebhook รับ Event จาก Stripe เมื่อมีการจ่ายเงินสำเร็จ
func (h *Handler) StripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusServiceUnavailable, "Error reading request body")
		return
	}

	// ใช้ Webhook Secret จาก Environment Variable (จากคำสั่ง stripe listen)
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	signatureHeader := c.GetHeader("Stripe-Signature")
	
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// เช็คว่าการจ่ายเงินสำเร็จ
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}

		// ดึง seat_id จาก Metadata ที่เราแนบไปตอนสร้าง Checkout Session
		seatIDStr := session.Metadata["seat_id"]
		seatID, err := strconv.ParseUint(seatIDStr, 10, 32)
		if err == nil {
			// เรียกใช้ Service เพื่อยืนยันการจอง (เปลี่ยนสถานะเป็น BOOKED)
			err = h.service.ConfirmBooking(context.Background(), uint(seatID))
			if err != nil {
				log.Printf("Failed to update seat status for SeatID %d: %v", seatID, err)
			} else {
				log.Printf("Successfully confirmed booking for SeatID %d", seatID)
			}
		} else {
			log.Printf("Invalid seat_id in metadata: %s", seatIDStr)
		}
	}

	// ส่ง 200 OK กลับไปให้ Stripe รู้ว่าเรารับทราบแล้ว
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}