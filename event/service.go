package event

import (
	"context"
)

type Service interface {
	GetEvents(ctx context.Context) ([]EventResponse, error)
	// 🔴 1. เพิ่มเมธอดดึงข้อมูล Event รายตัวเข้าใน Interface
	GetEventByID(ctx context.Context, id uint) (*EventResponse, error)
	GetSeats(ctx context.Context, eventID uint) ([]SeatResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetEvents(ctx context.Context) ([]EventResponse, error) {
	events, err := s.repo.FindAll(ctx)
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
		})
	}
	return res, nil
}

// 🔴 2. เพิ่ม Implementation ของ GetEventByID โดยเรียกใช้ Repo ตรงๆ และแมปค่าเป็น EventResponse
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
	}
	return res, nil
}

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