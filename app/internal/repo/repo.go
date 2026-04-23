// Package repo implements the data access layer of the application.
// It handles all database queries, transactions, and data mapping,
// providing a clean interface for the service layer to interact with PostgreSQL.
package repo

import (
	"context"

	"hotel.com/app/internal/models"
)

type ServiceRepository interface {
	Foo(ctx context.Context) error
	DbPing() error

	// Hotel methods
	ListHotels(ctx context.Context, city string, limit, offset int) ([]*models.Hotel, error)
	CreateHotel(ctx context.Context, hotel *models.Hotel) error
	GetHotelByID(ctx context.Context, id string) (*models.Hotel, error)
	UpdateHotel(ctx context.Context, hotel *models.Hotel) error
	DeleteHotel(ctx context.Context, id string) error

	// Review methods
	CreateReview(ctx context.Context, review *models.Review) error
	ListReviewsByHotelID(ctx context.Context, hotelID string, limit, offset int) ([]*models.Review, error)
	UpdateHotelRating(ctx context.Context, hotelID string) error
}

//REMEMBER TRANSACTION CODE LOGIC
