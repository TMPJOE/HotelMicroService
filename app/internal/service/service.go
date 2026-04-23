package service

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/google/uuid"
	"hotel.com/app/internal/client"
	"hotel.com/app/internal/models"
	"hotel.com/app/internal/repo"
)

type Service interface {
	Check() error

	// Hotel methods
	ListHotels(ctx context.Context, city string, limit, offset int) ([]*models.Hotel, error)
	CreateHotel(ctx context.Context, hotel *models.Hotel) error
	CreateHotelWithFiles(ctx context.Context, hotel *models.Hotel, files []models.FileUpload) error
	GetHotelByID(ctx context.Context, id string) (*models.Hotel, error)
	UpdateHotel(ctx context.Context, hotel *models.Hotel) error
	DeleteHotel(ctx context.Context, id string) error

	// Review methods
	CreateReview(ctx context.Context, review *models.Review) error
	ListReviewsByHotelID(ctx context.Context, hotelID string, limit, offset int) ([]*models.Review, error)
}

type fooService struct {
	l  *slog.Logger
	r  repo.ServiceRepository
	mc client.MediaClient
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

func (s *fooService) CreateHotelWithFiles(ctx context.Context, hotel *models.Hotel, files []models.FileUpload) error {
	hotel.ID = uuid.NewString()

	if err := s.r.CreateHotel(ctx, hotel); err != nil {
		return err
	}

	for _, file := range files {
		_, err := s.mc.UploadFile(ctx, bytes.NewReader(file.Content), file.Filename, "hotel", hotel.ID, file.ContentType)
		if err != nil {
			s.l.Error("failed to upload file to media service", "error", err, "filename", file.Filename)
			return err
		}
	}

	return nil
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
	return s.r.UpdateHotelRating(ctx, review.HotelID)
}

func (s *fooService) ListReviewsByHotelID(ctx context.Context, hotelID string, limit, offset int) ([]*models.Review, error) {
	return s.r.ListReviewsByHotelID(ctx, hotelID, limit, offset)
}

func New(l *slog.Logger, r repo.ServiceRepository, mc client.MediaClient) Service {
	return &fooService{
		l:  l,
		r:  r,
		mc: mc,
	}
}
