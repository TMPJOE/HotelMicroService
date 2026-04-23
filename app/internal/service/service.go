// Package service contains the business logic layer of the application.
// It defines service interfaces and implements use cases by orchestrating
// repositories, applying business rules, and returning results to handlers.
package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"hotel.com/app/internal/models"
	"hotel.com/app/internal/repo"
)

type Service interface {
	Check() error

	// Hotel methods
	ListHotels(ctx context.Context, city string, limit, offset int) ([]*models.Hotel, error)
	CreateHotel(ctx context.Context, hotel *models.Hotel) error
	GetHotelByID(ctx context.Context, id string) (*models.Hotel, error)
	UpdateHotel(ctx context.Context, hotel *models.Hotel) error
	DeleteHotel(ctx context.Context, id string) error

	// Review methods
	CreateReview(ctx context.Context, review *models.Review) error
	ListReviewsByHotelID(ctx context.Context, hotelID string, limit, offset int) ([]*models.Review, error)
}

type fooService struct {
	l *slog.Logger
	r repo.ServiceRepository
}

func (s *fooService) Check() error {
	s.l.Info("Pinging db...")
	err := s.r.DbPing()
	s.l.Info("is service working", "err", err.Error())
	return err
}

func (s *fooService) ListHotels(ctx context.Context, city string, limit, offset int) ([]*models.Hotel, error) {
	return s.r.ListHotels(ctx, city, limit, offset)
}

func (s *fooService) CreateHotel(ctx context.Context, hotel *models.Hotel) error {
	hotel.ID = uuid.NewString()
	return s.r.CreateHotel(ctx, hotel)
}

func (s *fooService) GetHotelByID(ctx context.Context, id string) (*models.Hotel, error) {
	return s.r.GetHotelByID(ctx, id)
}

func (s *fooService) UpdateHotel(ctx context.Context, hotel *models.Hotel) error {
	return s.r.UpdateHotel(ctx, hotel)
}

func (s *fooService) DeleteHotel(ctx context.Context, id string) error {
	return s.r.DeleteHotel(ctx, id)
}

func (s *fooService) CreateReview(ctx context.Context, review *models.Review) error {
	review.ID = uuid.NewString()
	err := s.r.CreateReview(ctx, review)
	if err != nil {
		return err
	}
	// Update the cached rating on the hotel asynchronously or synchronously
	// Going synchronous to guarantee consistency for immediate fetches
	return s.r.UpdateHotelRating(ctx, review.HotelID)
}

func (s *fooService) ListReviewsByHotelID(ctx context.Context, hotelID string, limit, offset int) ([]*models.Review, error) {
	return s.r.ListReviewsByHotelID(ctx, hotelID, limit, offset)
}

func New(l *slog.Logger, r repo.ServiceRepository) Service {
	return &fooService{
		l: l,
		r: r,
	}
}
