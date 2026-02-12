package postgres

import (
	"context"
	"database/sql"
	"strings"

	"ap_final/internal/domain"
	"ap_final/internal/store"
)

type MasterStore struct{ db *sql.DB }

func NewMasterStore(db *sql.DB) *MasterStore { return &MasterStore{db: db} }

func (s *MasterStore) ListBySalon(ctx context.Context, salonID string) ([]domain.Master, error) {
	const q = `
SELECT id, salon_id, full_name, bio, portfolio_url, is_active, created_at
FROM masters
WHERE salon_id = $1
ORDER BY full_name
`
	rows, err := s.db.QueryContext(ctx, q, salonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Master{}
	for rows.Next() {
		var m domain.Master
		var bio, port sql.NullString
		if err := rows.Scan(&m.ID, &m.SalonID, &m.FullName, &bio, &port, &m.IsActive, &m.CreatedAt); err != nil {
			return nil, err
		}
		if bio.Valid {
			m.Bio = bio.String
		}
		if port.Valid {
			m.PortfolioURL = port.String
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *MasterStore) GetByID(ctx context.Context, id string) (domain.Master, bool, error) {
	const q = `
SELECT id, salon_id, full_name, bio, portfolio_url, is_active, created_at
FROM masters WHERE id = $1
`
	var m domain.Master
	var bio, port sql.NullString
	err := s.db.QueryRowContext(ctx, q, id).Scan(&m.ID, &m.SalonID, &m.FullName, &bio, &port, &m.IsActive, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return domain.Master{}, false, nil
	}
	if err != nil {
		return domain.Master{}, false, err
	}
	if bio.Valid {
		m.Bio = bio.String
	}
	if port.Valid {
		m.PortfolioURL = port.String
	}
	return m, true, nil
}

// NEW: RBAC master needs user -> master binding
func (s *MasterStore) GetIDByUserID(ctx context.Context, userID string) (string, bool, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", false, nil
	}

	const q = `
SELECT id
FROM masters
WHERE user_id = $1 AND is_active = true
`
	var id string
	err := s.db.QueryRowContext(ctx, q, userID).Scan(&id)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (s *MasterStore) CreateForSalon(ctx context.Context, m domain.Master) (domain.Master, error) {
	const q = `
INSERT INTO masters (id, salon_id, full_name, bio, portfolio_url, is_active)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING created_at
`
	err := s.db.QueryRowContext(ctx, q,
		m.ID, m.SalonID, m.FullName, nullIfEmpty(m.Bio), nullIfEmpty(m.PortfolioURL), m.IsActive,
	).Scan(&m.CreatedAt)
	if err != nil {
		return domain.Master{}, mapPgError(err)
	}
	return m, nil
}

func (s *MasterStore) Update(ctx context.Context, m domain.Master) (domain.Master, error) {
	const q = `
UPDATE masters
SET full_name=$2, bio=$3, portfolio_url=$4, is_active=$5
WHERE id=$1
RETURNING salon_id, created_at
`
	err := s.db.QueryRowContext(ctx, q,
		m.ID, m.FullName, nullIfEmpty(m.Bio), nullIfEmpty(m.PortfolioURL), m.IsActive,
	).Scan(&m.SalonID, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return domain.Master{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Master{}, mapPgError(err)
	}
	return m, nil
}

func (s *MasterStore) Disable(ctx context.Context, id string) error {
	const q = `UPDATE masters SET is_active=false WHERE id=$1`
	res, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return mapPgError(err)
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return store.ErrNotFound
	}
	return nil
}
