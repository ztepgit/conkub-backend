// event/dto.go
package event

import "time"

// EventResponse เป็นรูปแบบข้อมูลที่ส่งให้ Next.js (หน้าแสดงคอนเสิร์ต)
type EventResponse struct {
	ID                uint      `json:"id"`
	Name              string    `json:"name"`
	Artist            string    `json:"artist"`
	Description       string    `json:"description"`
	Venue             string    `json:"venue"`
	Category          string    `json:"category"`
	ImageURL          string    `json:"image_url"`
	ShowTime          time.Time `json:"show_time"`
	RemainingTickets  int64     `json:"remaining_tickets"`
}

// SeatResponse เป็นรูปแบบข้อมูลที่นั่ง
type SeatResponse struct {
	ID     uint    `json:"id"`
	Row    string  `json:"row"`
	Number int     `json:"number"`
	Price  float64 `json:"price"`
	Status string  `json:"status"` // "AVAILABLE" หรือ "BOOKED"
}