// Package handler provides HTTP request handlers, routing, and middleware.
// It handles incoming HTTP requests, delegates to the service layer for
// business logic, and returns JSON responses with appropriate status codes.
package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"hotel.com/app/internal/helper"
	"hotel.com/app/internal/models"
	"hotel.com/app/internal/service"
)

type Handler struct {
	s       service.Service
	l       *slog.Logger
	jwtAuth *JWTAuthenticator
}

func New(s service.Service, l *slog.Logger, jwtAuth *JWTAuthenticator) *Handler {
	return &Handler{
		s:       s,
		l:       l,
		jwtAuth: jwtAuth,
	}
}

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// readinessCheck verifies if the service is ready to accept traffic
// by pinging the database and other critical dependencies.
func (h *Handler) readinessCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Check(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "reason": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
		"db":     "ok",
	})
}

func (h *Handler) handleListHotels(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	hotels, err := h.s.ListHotels(r.Context(), city, limit, offset)
	if err != nil {
		h.l.Error("failed to list hotels", "error", err)
		helper.RespondError(w, http.StatusInternalServerError, "failed to fetch hotels")
		return
	}

	if hotels == nil {
		hotels = []*models.Hotel{}
	}

	helper.RespondJSON(w, http.StatusOK, hotels)
}

func (h *Handler) handleCreateHotel(w http.ResponseWriter, r *http.Request) {
	claims := GetClaimsFromRequest(r)
	if claims == nil {
		helper.RespondError(w, http.StatusUnauthorized, helper.ErrUnauthorized.Error())
		return
	}
	if !strings.EqualFold(claims.UserType, "admin") {
		helper.RespondError(w, http.StatusForbidden, "your account does not have admin privileges")
		return
	}

	var hotel *models.Hotel
	var files []models.FileUpload

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			helper.RespondError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		hotel = &models.Hotel{
			AdminID: claims.UserID,
			Name:    r.FormValue("name"),
			City:    r.FormValue("city"),
		}

		if desc := r.FormValue("description"); desc != "" {
			hotel.Description = desc
		}
		if lat := r.FormValue("lat"); lat != "" {
			if parsed, err := strconv.ParseFloat(lat, 64); err == nil {
				hotel.Lat = parsed
			}
		}
		if lng := r.FormValue("lng"); lng != "" {
			if parsed, err := strconv.ParseFloat(lng, 64); err == nil {
				hotel.Lng = parsed
			}
		}

		if r.MultipartForm != nil {
			for _, fileHeaders := range r.MultipartForm.File {
				for _, header := range fileHeaders {
					file, err := header.Open()
					if err != nil {
						h.l.Error("failed to open uploaded file", "error", err)
						helper.RespondError(w, http.StatusBadRequest, "failed to open uploaded file")
						return
					}
					defer file.Close()

					content := make([]byte, header.Size)
					if _, err := io.ReadFull(file, content); err != nil {
						h.l.Error("failed to read uploaded file", "error", err)
						helper.RespondError(w, http.StatusBadRequest, "failed to read uploaded file")
						return
					}

					files = append(files, models.FileUpload{
						Filename:    header.Filename,
						Content:     content,
						ContentType: header.Header.Get("Content-Type"),
					})
				}
			}
		}
	} else {
		var req models.CreateHotelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.RespondError(w, http.StatusBadRequest, "invalid request payload")
			return
		}

		if req.Name == "" || req.City == "" {
			helper.RespondError(w, http.StatusBadRequest, "name and city are required")
			return
		}

		hotel = &models.Hotel{
			AdminID:     claims.UserID,
			Name:        req.Name,
			City:        req.City,
			Description: req.Description,
			Lat:         req.Lat,
			Lng:         req.Lng,
		}
	}

	if hotel.Name == "" || hotel.City == "" {
		helper.RespondError(w, http.StatusBadRequest, "name and city are required")
		return
	}

	var err error
	if len(files) > 0 {
		err = h.s.CreateHotelWithFiles(r.Context(), hotel, files)
	} else {
		err = h.s.CreateHotel(r.Context(), hotel)
	}

	if err != nil {
		h.l.Error("failed to create hotel", "error", err)
		helper.RespondError(w, http.StatusInternalServerError, "failed to create hotel")
		return
	}

	helper.RespondJSON(w, http.StatusCreated, hotel)
}

func (h *Handler) handleGetHotel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	hotel, err := h.s.GetHotelByID(r.Context(), id)
	if err != nil {
		helper.RespondError(w, http.StatusNotFound, "hotel not found")
		return
	}
	helper.RespondJSON(w, http.StatusOK, hotel)
}

func (h *Handler) handleUpdateHotel(w http.ResponseWriter, r *http.Request) {
	claims := GetClaimsFromRequest(r)
	if claims == nil {
		helper.RespondError(w, http.StatusUnauthorized, helper.ErrUnauthorized.Error())
		return
	}
	if !strings.EqualFold(claims.UserType, "admin") {
		helper.RespondError(w, http.StatusForbidden, "your account does not have admin privileges")
		return
	}

	id := chi.URLParam(r, "id")
	hotel, err := h.s.GetHotelByID(r.Context(), id)
	if err != nil {
		helper.RespondError(w, http.StatusNotFound, "hotel not found")
		return
	}

	// Make sure the admin updating is the owner of the hotel
	if hotel.AdminID != claims.UserID {
		helper.RespondError(w, http.StatusForbidden, "you can only update your own hotels")
		return
	}

	var req models.UpdateHotelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if req.Name != nil {
		hotel.Name = *req.Name
	}
	if req.City != nil {
		hotel.City = *req.City
	}
	if req.Description != nil {
		hotel.Description = *req.Description
	}
	if req.Lat != nil {
		hotel.Lat = *req.Lat
	}
	if req.Lng != nil {
		hotel.Lng = *req.Lng
	}

	if err := h.s.UpdateHotel(r.Context(), hotel); err != nil {
		h.l.Error("failed to update hotel", "error", err)
		helper.RespondError(w, http.StatusInternalServerError, "failed to update hotel")
		return
	}

	helper.RespondJSON(w, http.StatusOK, hotel)
}

func (h *Handler) handleDeleteHotel(w http.ResponseWriter, r *http.Request) {
	claims := GetClaimsFromRequest(r)
	if claims == nil {
		helper.RespondError(w, http.StatusUnauthorized, helper.ErrUnauthorized.Error())
		return
	}
	if !strings.EqualFold(claims.UserType, "admin") {
		helper.RespondError(w, http.StatusForbidden, "your account does not have admin privileges")
		return
	}

	id := chi.URLParam(r, "id")
	hotel, err := h.s.GetHotelByID(r.Context(), id)
	if err != nil {
		helper.RespondError(w, http.StatusNotFound, "hotel not found")
		return
	}

	if hotel.AdminID != claims.UserID {
		helper.RespondError(w, http.StatusForbidden, "you can only delete your own hotels")
		return
	}

	if err := h.s.DeleteHotel(r.Context(), id); err != nil {
		h.l.Error("failed to delete hotel", "error", err)
		helper.RespondError(w, http.StatusInternalServerError, "failed to delete hotel")
		return
	}

	helper.RespondJSON(w, http.StatusOK, map[string]string{"message": "hotel deleted successfully"})
}

func (h *Handler) handleListReviews(w http.ResponseWriter, r *http.Request) {
	hotelID := chi.URLParam(r, "id")

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	reviews, err := h.s.ListReviewsByHotelID(r.Context(), hotelID, limit, offset)
	if err != nil {
		h.l.Error("failed to list reviews", "error", err)
		helper.RespondError(w, http.StatusInternalServerError, "failed to fetch reviews")
		return
	}

	if reviews == nil {
		reviews = []*models.Review{}
	}

	helper.RespondJSON(w, http.StatusOK, reviews)
}

func (h *Handler) handleCreateReview(w http.ResponseWriter, r *http.Request) {
	claims := GetClaimsFromRequest(r)
	if claims == nil {
		helper.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	hotelID := chi.URLParam(r, "id")

	// Check if hotel exists
	_, err := h.s.GetHotelByID(r.Context(), hotelID)
	if err != nil {
		helper.RespondError(w, http.StatusNotFound, "hotel not found")
		return
	}

	var req models.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.RespondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		helper.RespondError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	review := &models.Review{
		HotelID: hotelID,
		UserID:  claims.UserID,
		Rating:  req.Rating,
		Comment: req.Comment,
	}

	if err := h.s.CreateReview(r.Context(), review); err != nil {
		h.l.Error("failed to create review", "error", err)
		helper.RespondError(w, http.StatusInternalServerError, "failed to create review (you might have already reviewed this hotel)")
		return
	}

	helper.RespondJSON(w, http.StatusCreated, review)
}
