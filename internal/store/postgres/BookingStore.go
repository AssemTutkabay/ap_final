package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"ap_final/internal/domain"
	"ap_final/internal/store"
)

type BookingStore struct{ db *sql.DB }

func NewBookingStore(db *sql.DB) *BookingStore { return &BookingStore{db: db} }

func (s *BookingStore) Create(ctx context.Context, b domain.Booking) (domain.Booking, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return domain.Booking{}, err
	}
	defer func() { _ = tx.Rollback() }()

	// lock per master to avoid concurrent overlap insert race
	const lockQ = `SELECT pg_advisory_xact_lock(hashtext($1))`
	if _, err := tx.ExecContext(ctx, lockQ, b.MasterID); err != nil {
		return domain.Booking{}, err
	}

	// salon_id by master_id
	var salonID string
	err = tx.QueryRowContext(ctx, `SELECT salon_id FROM masters WHERE id = $1`, b.MasterID).Scan(&salonID)
	if err == sql.ErrNoRows {
		return domain.Booking{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Booking{}, err
	}

	// service must belong to same salon
	var one int
	err = tx.QueryRowContext(ctx,
		`SELECT 1 FROM services WHERE id=$1 AND salon_id=$2 AND is_active=true`,
		b.ServiceID, salonID,
	).Scan(&one)
	if err == sql.ErrNoRows {
		return domain.Booking{}, store.ErrConflict
	}
	if err != nil {
		return domain.Booking{}, err
	}

	// overlap check
	err = tx.QueryRowContext(ctx, `
SELECT 1
FROM bookings
WHERE master_id = $1
  AND status IN ('created','confirmed')
  AND $2 < end_at
  AND $3 > start_at
LIMIT 1
`, b.MasterID, b.StartAt, b.EndAt).Scan(&one)
	if err == nil {
		return domain.Booking{}, store.ErrConflict
	}
	if err != sql.ErrNoRows {
		return domain.Booking{}, err
	}

	// insert
	const insQ = `
INSERT INTO bookings (salon_id, client_id, master_id, service_id, start_at, end_at, status, cancel_reason)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING id, created_at
`
	var id int64
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, insQ,
		salonID,
		b.ClientID,
		b.MasterID,
		b.ServiceID,
		b.StartAt,
		b.EndAt,
		string(b.Status),
		nullIfEmpty(b.CancelReason),
	).Scan(&id, &createdAt)
	if err != nil {
		return domain.Booking{}, mapPgError(err)
	}

	if err := tx.Commit(); err != nil {
		return domain.Booking{}, err
	}

	b.ID = strconv.FormatInt(id, 10)
	b.SalonID = salonID
	b.CreatedAt = createdAt
	return b, nil
}

func (s *BookingStore) List(ctx context.Context, f store.BookingListFilter) ([]domain.Booking, error) {
	q := `
SELECT id, salon_id, client_id, master_id, service_id, start_at, end_at, status, cancel_reason, created_at
FROM bookings
WHERE 1=1
`
	args := []any{}
	n := 1

	if strings.TrimSpace(f.ClientID) != "" {
		q += " AND client_id = $" + strconv.Itoa(n)
		args = append(args, f.ClientID)
		n++
	}
	if strings.TrimSpace(f.MasterID) != "" {
		q += " AND master_id = $" + strconv.Itoa(n)
		args = append(args, f.MasterID)
		n++
	}
	if strings.TrimSpace(f.SalonID) != "" {
		q += " AND salon_id = $" + strconv.Itoa(n)
		args = append(args, f.SalonID)
		n++
	}

	q += " ORDER BY start_at DESC"

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Booking{}
	for rows.Next() {
		var b domain.Booking
		var id int64
		var status string
		var cancel sql.NullString

		if err := rows.Scan(
			&id, &b.SalonID, &b.ClientID, &b.MasterID, &b.ServiceID,
			&b.StartAt, &b.EndAt, &status, &cancel, &b.CreatedAt,
		); err != nil {
			return nil, err
		}
		b.ID = strconv.FormatInt(id, 10)
		b.Status = domain.BookingStatus(status)
		if cancel.Valid {
			b.CancelReason = cancel.String
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *BookingStore) GetByID(ctx context.Context, id string) (domain.Booking, bool, error) {
	bid, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || bid <= 0 {
		return domain.Booking{}, false, nil
	}

	const q = `
SELECT id, salon_id, client_id, master_id, service_id, start_at, end_at, status, cancel_reason, created_at
FROM bookings
WHERE id = $1
`
	var b domain.Booking
	var rid int64
	var status string
	var cancel sql.NullString

	err = s.db.QueryRowContext(ctx, q, bid).Scan(
		&rid, &b.SalonID, &b.ClientID, &b.MasterID, &b.ServiceID,
		&b.StartAt, &b.EndAt, &status, &cancel, &b.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Booking{}, false, nil
	}
	if err != nil {
		return domain.Booking{}, false, err
	}

	b.ID = strconv.FormatInt(rid, 10)
	b.Status = domain.BookingStatus(status)
	if cancel.Valid {
		b.CancelReason = cancel.String
	}
	return b, true, nil
}

func (s *BookingStore) CancelByClient(ctx context.Context, bookingID, clientID string, cancelReason string, isLateCancel bool) (domain.Booking, error) {
	bid, err := strconv.ParseInt(strings.TrimSpace(bookingID), 10, 64)
	if err != nil || bid <= 0 {
		return domain.Booking{}, store.ErrNotFound
	}

	// Пытаемся отменить только active.
	const q = `
UPDATE bookings
SET status='cancelled_by_client', cancel_reason=$3, updated_at=now()
WHERE id=$1 AND client_id=$2 AND status IN ('created','confirmed')
RETURNING id, salon_id, client_id, master_id, service_id, start_at, end_at, status, cancel_reason, created_at
`
	var b domain.Booking
	var rid int64
	var status string
	var cancel sql.NullString

	err = s.db.QueryRowContext(ctx, q, bid, clientID, nullIfEmpty(cancelReason)).Scan(
		&rid, &b.SalonID, &b.ClientID, &b.MasterID, &b.ServiceID,
		&b.StartAt, &b.EndAt, &status, &cancel, &b.CreatedAt,
	)
	if err == nil {
		b.ID = strconv.FormatInt(rid, 10)
		b.Status = domain.BookingStatus(status)
		if cancel.Valid {
			b.CancelReason = cancel.String
		}
		return b, nil
	}

	// Если не обновили, проверяем почему.
	existing, ok, gerr := s.GetByID(ctx, bookingID)
	if gerr != nil {
		return domain.Booking{}, gerr
	}
	if !ok {
		return domain.Booking{}, store.ErrNotFound
	}
	if existing.ClientID != clientID {
		return domain.Booking{}, store.ErrForbidden
	}
	// Уже не active (cancelled/completed) => возвращаем как есть.
	return existing, nil
}

func (s *BookingStore) UpdateStatus(ctx context.Context, bookingID string, newStatus domain.BookingStatus, cancelReason string) (domain.Booking, error) {
	bid, err := strconv.ParseInt(strings.TrimSpace(bookingID), 10, 64)
	if err != nil || bid <= 0 {
		return domain.Booking{}, store.ErrNotFound
	}

	const q = `
UPDATE bookings
SET status=$2, cancel_reason=$3, updated_at=now()
WHERE id=$1
RETURNING id, salon_id, client_id, master_id, service_id, start_at, end_at, status, cancel_reason, created_at
`
	var b domain.Booking
	var rid int64
	var status string
	var cancel sql.NullString

	err = s.db.QueryRowContext(ctx, q, bid, string(newStatus), nullIfEmpty(cancelReason)).Scan(
		&rid, &b.SalonID, &b.ClientID, &b.MasterID, &b.ServiceID,
		&b.StartAt, &b.EndAt, &status, &cancel, &b.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Booking{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Booking{}, mapPgError(err)
	}

	b.ID = strconv.FormatInt(rid, 10)
	b.Status = domain.BookingStatus(status)
	if cancel.Valid {
		b.CancelReason = cancel.String
	}
	return b, nil
}
