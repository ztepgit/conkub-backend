package db

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitPostgres(dsn string) *gorm.DB {
	// 🔴 1. ปรับแต่ง Postgres Driver ให้เข้ากันได้กับ PgBouncer
	pgConfig := postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // บังคับไม่ให้ pgx driver สร้าง prepared statements แอบแฝง
	}

	// 🔴 2. ตั้งค่า GORM โดยรวมของเดิมและของใหม่เข้าด้วยกัน
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,                                // คงของเดิมไว้: ปิด default transaction เพื่อให้ควบคุมเองได้ 100%
		PrepareStmt:            false,                               // ปิดการทำ Statement Caching แก้ปัญหา SQLSTATE 42P05
		Logger:                 logger.Default.LogMode(logger.Info), // เปิด Info ดู Query เพื่อยืนยันว่าไม่ได้ใช้ stmtcache
	}

	db, err := gorm.Open(postgres.New(pgConfig), gormConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// 🔴 3. Production-ready connection pool settings (รวมการตั้งค่าสำหรับ PgBouncer)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // เพิ่มค่า MaxIdleTime ให้เหมาะสมกับ Connection Pool

	log.Println("✅ PostgreSQL connected successfully with PgBouncer compatibility")
	return db
}