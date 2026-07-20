// event/dto.go
package event

import "time"

// EventResponse เป็นรูปแบบข้อมูลที่ส่งให้ Next.js (หน้าแสดงคอนเสิร์ต)
type EventResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Venue       string    `json:"venue"`
	ShowTime    time.Time `json:"show_time"`
}

// SeatResponse เป็นรูปแบบข้อมูลที่นั่ง
type SeatResponse struct {
	ID     uint    `json:"id"`
	Row    string  `json:"row"`
	Number int     `json:"number"`
	Price  float64 `json:"price"`
	Status string  `json:"status"` // "AVAILABLE" หรือ "BOOKED"
}