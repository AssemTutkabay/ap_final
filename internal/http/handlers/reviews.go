package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ap_final/internal/domain"
	"ap_final/internal/http/middleware"
	"ap_final/internal/store"
)

type ReviewsHandler struct {
	reviews  store.ReviewStore
	bookings store.BookingStore
}

func NewReviewsHandler(r store.ReviewStore, b store.BookingStore) *ReviewsHandler {
	return &ReviewsHandler{reviews: r, bookings: b}
}

type CreateReviewRequest struct {
	BookingID string `json:"bookingId"`
	Rating    int    `json:"rating"`
	Text      string `json:"text,omitempty"`
}

func (h *ReviewsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	u, ok := middleware.UserFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Роутер уже ограничил роль client, но пусть будет и тут защитно.
	if u.Role != domain.RoleClient {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}

	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}
	if req.BookingID == "" {
		writeErr(w, http.StatusBadRequest, "bookingId required")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		writeErr(w, http.StatusBadRequest, "rating must be 1..5")
		return
	}

	b, ok2, err := h.bookings.GetByID(r.Context(), req.BookingID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !ok2 {
		writeErr(w, http.StatusNotFound, "booking not found")
		return
	}

	// client ownership
	if b.ClientID != u.ID {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}

	if b.Status != domain.BookingCompleted {
		writeErr(w, http.StatusBadRequest, "booking must be completed")
		return
	}

	// 1 booking = 1 review (verified)
	_, exists, err := h.reviews.GetByBookingID(r.Context(), req.BookingID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if exists {
		writeErr(w, http.StatusConflict, "conflict: review already exists")
		return
	}

	newR := domain.Review{
		BookingID: req.BookingID,
		ClientID:  u.ID, // берем из токена
		SalonID:   b.SalonID,
		MasterID:  b.MasterID,
		Rating:    req.Rating,
		Text:      req.Text,
	}

	created, err := h.reviews.Create(r.Context(), newR)
	if err != nil {
		if strings.Contains(err.Error(), "conflict") {
			writeErr(w, http.StatusConflict, "conflict: cannot create review")
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// /salons/{id}/reviews (public)
func (h *ReviewsHandler) HandleBySalon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/salons/")
	if !strings.HasSuffix(path, "/reviews") {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	id := strings.TrimSuffix(path, "/reviews")
	id = strings.TrimSuffix(id, "/")
	if id == "" || strings.Contains(id, "/") {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}

	list, err := h.reviews.ListBySalon(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, list)
}
