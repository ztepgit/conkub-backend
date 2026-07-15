package booking

import (
	"context"
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

	err := h.service.BookSeat(ctx, userID, req)
	if err != nil {
		// จัดการ Error Message ตามที่ Service โยนมา
		if err.Error() == "seat already booked" || err.Error() == "seat is currently being booked by someone else" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // HTTP 409 Conflict
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "seat booked successfully"})
}