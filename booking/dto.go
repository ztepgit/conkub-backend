package booking

// BookSeatRequest เป็นรูปแบบ JSON ที่ Next.js จะส่งมาตอนกดจอง
type BookSeatRequest struct {
	EventID uint `json:"event_id" binding:"required"`
	SeatID  uint `json:"seat_id" binding:"required"`
}