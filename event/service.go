// event/service.go
package event

import (
	"context"
	"time"
)

type Service interface {
	GetEvents(ctx context.Context, search, location string, parsedDate *time.Time) ([]EventResponse, error)
	GetEventByID(ctx context.Context, id uint) (*EventResponse, error)
	GetSeats(ctx context.Context, eventID uint) ([]SeatResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetEvents(ctx context.Context, search, location string, parsedDate *time.Time) ([]EventResponse, error) {
	// 🔴 ส่งผ่านพารามิเตอร์ค้นหาไปยัง Repository
	events, err := s.repo.FindAll(ctx, search, location, parsedDate)
	if err != nil {
		return nil, err
	}

	var res []EventResponse
	for _, e := range events {
		res = append(res, EventResponse{
			ID:               e.ID,
			Name:             e.Name,
			Artist:           e.Artist,
			Description:      e.Description,
			Venue:            e.Venue,
			Category:         e.Category,
			ImageURL:         e.ImageURL,
			ShowTime:         e.ShowTime,
			RemainingTickets: e.RemainingTickets,
			Price:            e.Price, // 🔴 คงการแมปข้อมูล Price จาก Repository สู่ DTO ไว้
		})
	}
	return res, nil
}

func (s *service) GetEventByID(ctx context.Context, id uint) (*EventResponse, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := &EventResponse{
		ID:               e.ID,
		Name:             e.Name,
		Artist:           e.Artist,
		Description:      e.Description,
		Venue:            e.Venue,
		Category:         e.Category,
		ImageURL:         e.ImageURL,
		ShowTime:         e.ShowTime,
		RemainingTickets: e.RemainingTickets,
		Price:            e.Price, // 🔴 คงการแมปข้อมูล Price จาก Repository สู่ DTO ไว้
	}
	return res, nil
}

// ฟังก์ชัน GetSeats คงไว้ตามเดิม
func (s *service) GetSeats(ctx context.Context, eventID uint) ([]SeatResponse, error) {
	seats, err := s.repo.FindSeatsByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	var res []SeatResponse
	for _, st := range seats {
		res = append(res, SeatResponse{
			ID:     st.ID,
			Row:    st.Row,
			Number: st.Number,
			Price:  st.Price,
			Status: string(st.Status),
		})
	}
	return res, nil
}