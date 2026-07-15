package main

import (
	"log"

	"conkub-backend/booking"    // เพิ่ม Import booking module
	"conkub-backend/config"
	"conkub-backend/db"
	"conkub-backend/event"      // Import event module
	"conkub-backend/middleware" // เพิ่ม Import middleware module

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Database & Redis
	database := db.InitPostgres(cfg.DatabaseURL)
	redisClient := db.InitRedis(cfg.RedisURL, cfg.RedisPassword)

	// --- 3. Modules Setup ---
	// Event Module Setup
	eventRepo := event.NewRepository(database)
	eventService := event.NewService(eventRepo)
	eventHandler := event.NewHandler(eventService)

	// Booking Module Setup
	bookingRepo := booking.NewRepository(database)
	bookingService := booking.NewService(bookingRepo, redisClient)
	bookingHandler := booking.NewHandler(bookingService)

	// 4. Setup Gin Router
	r := gin.Default()

	// Public Routes (ไม่ต้องใช้ Middleware)
	api := r.Group("/api/v1")
	{
		// ใช้เช็คสถานะ Server
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		api.GET("/events", eventHandler.GetEvents)
		api.GET("/events/:id/seats", eventHandler.GetSeats)
		
		// 🔴 เพิ่ม Endpoint สำหรับ Stripe Webhook (ต้องเป็น Public)
		api.POST("/webhook/stripe", bookingHandler.StripeWebhook)
	}

	// Protected Routes (ต้องล็อกอินและใช้ JWT Middleware)
	protected := r.Group("/api/v1")
	protected.Use(middleware.RequireAuth(cfg.SupabaseJWTSecret))
	{
		// หน้าบ้านจะต้องส่ง Header -> Authorization: Bearer <Supabase_Token>
		protected.POST("/bookings", bookingHandler.BookSeat)
	}

	// 5. Start Server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}