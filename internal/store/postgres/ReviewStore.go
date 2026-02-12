package postgres

import (
	"context"
	"database/sql"
	"strconv"

	"ap_final/internal/domain"
)

type ReviewStore struct{ db *sql.DB }

func NewReviewStore(db *sql.DB) *ReviewStore { return &ReviewStore{db: db} }

func (s *ReviewStore) Create(ctx context.Context, r domain.Review) (domain.Review, error) {
	const q = `
INSERT INTO reviews (booking_id, client_id, salon_id, master_id, rating, text)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id, created_at
`
	var id int64
	err := s.db.QueryRowContext(ctx, q,
		r.BookingID, r.ClientID, r.SalonID, r.MasterID, r.Rating, nullIfEmpty(r.Text),
	).Scan(&id, &r.CreatedAt)

	if err != nil {
		return domain.Review{}, mapPgError(err)
	}
	r.ID = strconv.FormatInt(id, 10)
	return r, nil
}

func (s *ReviewStore) ListBySalon(ctx context.Context, salonID string) ([]domain.Review, error) {
	const q = `
SELECT id, booking_id, client_id, salon_id, master_id, rating, text, created_at
FROM reviews
WHERE salon_id = $1
ORDER BY created_at DESC
`
	rows, err := s.db.QueryContext(ctx, q, salonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Review{}
	for rows.Next() {
		var r domain.Review
		var id int64
		var text sql.NullString
		if err := rows.Scan(&id, &r.BookingID, &r.ClientID, &r.SalonID, &r.MasterID, &r.Rating, &text, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.ID = strconv.FormatInt(id, 10)
		if text.Valid {
			r.Text = text.String
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *ReviewStore) GetByBookingID(ctx context.Context, bookingID string) (domain.Review, bool, error) {
	const q = `
SELECT id, booking_id, client_id, salon_id, master_id, rating, text, created_at
FROM reviews
WHERE booking_id = $1
`
	var r domain.Review
	var id int64
	var text sql.NullString

	err := s.db.QueryRowContext(ctx, q, bookingID).Scan(&id, &r.BookingID, &r.ClientID, &r.SalonID, &r.MasterID, &r.Rating, &text, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return domain.Review{}, false, nil
	}
	if err != nil {
		return domain.Review{}, false, err
	}
	r.ID = strconv.FormatInt(id, 10)
	if text.Valid {
		r.Text = text.String
	}
	return r, true, nil
}
