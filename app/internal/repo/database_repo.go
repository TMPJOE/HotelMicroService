package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"hotel.com/app/internal/models"
)

type databaseRepo struct {
	db *pgxpool.Pool
}

func NewDatabaseRepo(conn *pgxpool.Pool) ServiceRepository {
	return &databaseRepo{
		db: conn,
	}
}

func (dbr *databaseRepo) DbPing() error {
	err := dbr.db.Ping(context.Background())
	return err
}

// ListHotels lists hotels, optionally filtering by city
func (dbr *databaseRepo) ListHotels(ctx context.Context, city string, limit, offset int) ([]*models.Hotel, error) {
	query := `SELECT id, admin_id, name, city, description, rating, lat, lng, created_at, updated_at FROM hotels`
	var args []interface{}
	argID := 1

	if city != "" {
		query += fmt.Sprintf(" WHERE city = $%d", argID)
		args = append(args, city)
		argID++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	rows, err := dbr.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotels []*models.Hotel
	for rows.Next() {
		h := &models.Hotel{}
		if err := rows.Scan(&h.ID, &h.AdminID, &h.Name, &h.City, &h.Description, &h.Rating, &h.Lat, &h.Lng, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		hotels = append(hotels, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hotels, nil
}

// CreateHotel creates a new hotel
func (dbr *databaseRepo) CreateHotel(ctx context.Context, hotel *models.Hotel) error {
	query := `INSERT INTO hotels (id, admin_id, name, city, description, lat, lng)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := dbr.db.Exec(ctx, query, hotel.ID, hotel.AdminID, hotel.Name, hotel.City, hotel.Description, hotel.Lat, hotel.Lng)
	return err
}

// GetHotelByID fetches a single hotel by ID
func (dbr *databaseRepo) GetHotelByID(ctx context.Context, id string) (*models.Hotel, error) {
	query := `SELECT id, admin_id, name, city, description, rating, lat, lng, created_at, updated_at FROM hotels WHERE id = $1`
	h := &models.Hotel{}
	err := dbr.db.QueryRow(ctx, query, id).Scan(
		&h.ID, &h.AdminID, &h.Name, &h.City, &h.Description, &h.Rating, &h.Lat, &h.Lng, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// UpdateHotel updates a hotel's information
func (dbr *databaseRepo) UpdateHotel(ctx context.Context, hotel *models.Hotel) error {
	query := `UPDATE hotels SET name = $1, city = $2, description = $3, lat = $4, lng = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $6`
	_, err := dbr.db.Exec(ctx, query, hotel.Name, hotel.City, hotel.Description, hotel.Lat, hotel.Lng, hotel.ID)
	return err
}

// DeleteHotel deletes a hotel by ID
func (dbr *databaseRepo) DeleteHotel(ctx context.Context, id string) error {
	query := `DELETE FROM hotels WHERE id = $1`
	_, err := dbr.db.Exec(ctx, query, id)
	return err
}

// CreateReview creates a new review
func (dbr *databaseRepo) CreateReview(ctx context.Context, review *models.Review) error {
	query := `INSERT INTO reviews (id, hotel_id, user_id, rating, comment) VALUES ($1, $2, $3, $4, $5)`
	_, err := dbr.db.Exec(ctx, query, review.ID, review.HotelID, review.UserID, review.Rating, review.Comment)
	return err
}

// ListReviewsByHotelID fetches reviews for a specific hotel
func (dbr *databaseRepo) ListReviewsByHotelID(ctx context.Context, hotelID string, limit, offset int) ([]*models.Review, error) {
	query := `SELECT id, hotel_id, user_id, rating, comment, created_at FROM reviews WHERE hotel_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := dbr.db.Query(ctx, query, hotelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*models.Review
	for rows.Next() {
		r := &models.Review{}
		if err := rows.Scan(&r.ID, &r.HotelID, &r.UserID, &r.Rating, &r.Comment, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return reviews, nil
}

// UpdateHotelRating updates the average rating of a hotel based on its reviews
func (dbr *databaseRepo) UpdateHotelRating(ctx context.Context, hotelID string) error {
	query := `UPDATE hotels SET rating = (SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE hotel_id = $1) WHERE id = $1`
	_, err := dbr.db.Exec(ctx, query, hotelID)
	return err
}
