// Package models defines domain data structures used across the application.
// All entities, DTOs, and shared types should be defined here
// to ensure consistency between repository, service, and handler layers.
package models

import "time"

type Hotel struct {
	ID          string    `json:"id" db:"id"`
	AdminID     string    `json:"admin_id" db:"admin_id"`
	Name        string    `json:"name" db:"name"`
	City        string    `json:"city" db:"city"`
	Description string    `json:"description" db:"description"`
	Rating      float64   `json:"rating" db:"rating"`
	Lat         float64   `json:"lat" db:"lat"`
	Lng         float64   `json:"lng" db:"lng"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Review struct {
	ID        string    `json:"id" db:"id"`
	HotelID   string    `json:"hotel_id" db:"hotel_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Request DTOs
type CreateHotelRequest struct {
	Name        string  `json:"name" validate:"required"`
	City        string  `json:"city" validate:"required"`
	Description string  `json:"description"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

type CreateHotelWithFilesRequest struct {
	Name        string
	City        string
	Description string
	Lat         float64
	Lng         float64
	Files       []FileUpload
}

type FileUpload struct {
	Filename    string
	Content     []byte
	ContentType string
}

type UpdateHotelRequest struct {
	Name        *string  `json:"name"`
	City        *string  `json:"city"`
	Description *string  `json:"description"`
	Lat         *float64 `json:"lat"`
	Lng         *float64 `json:"lng"`
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment"`
}
