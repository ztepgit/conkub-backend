// event/handler.go
package event

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler interface {
	GetEvents(c *gin.Context)
	GetEventByID(c *gin.Context)
	GetSeats(c *gin.Context)
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

func (h *handler) GetEvents(c *gin.Context) {
	// ดึง Query Parameters
	search := c.Query("search")
	location := c.Query("location")
	dateStr := c.Query("date")

	var parsedDate *time.Time

	// Parse วันที่ให้อยู่ใน Timezone Asia/Bangkok
	if dateStr != "" {
		loc, err := time.LoadLocation("Asia/Bangkok")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load timezone"})
			return
		}

		t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, expected YYYY-MM-DD"})
			return
		}
		parsedDate = &t
	}

	// สร้าง Context timeout เพื่อกัน API ค้าง
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// เรียกใช้งาน Service พร้อมส่ง Parameter ค้นหา
	events, err := h.service.GetEvents(ctx, search, location, parsedDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// คืนค่าเป็น Data เปล่าแทน Null หากไม่พบข้อมูล
	if events == nil {
		events = []EventResponse{}
	}

	c.JSON(http.StatusOK, gin.H{"data": events})
}

// 🔴 คงฟังก์ชัน GetEventByID สำหรับรองรับ Route: GET /api/v1/events/:id
func (h *handler) GetEventByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	// ใช้ Context timeout 5 วินาทีตามรูปแบบมาตรฐานของโปรเจกต์
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	event, err := h.service.GetEventByID(ctx, uint(id))
	if err != nil {
		// ดักจับ Error จาก GORM หากไม่พบข้อมูลให้คืนค่า 404 Not Found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// ตอบกลับด้วยโครงสร้างเดียวกับ GetEvents และ GetSeats
	c.JSON(http.StatusOK, gin.H{"data": event})
}

// 🔴 คงฟังก์ชัน GetSeats ไว้เหมือนเดิม
func (h *handler) GetSeats(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	seats, err := h.service.GetSeats(ctx, uint(eventID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch seats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": seats})
}