package postgres

import (
	"context"
	"database/sql"

	"ap_final/internal/domain"
	"ap_final/internal/store"
)

type ServiceStore struct{ db *sql.DB }

func NewServiceStore(db *sql.DB) *ServiceStore { return &ServiceStore{db: db} }

func (s *ServiceStore) ListBySalon(ctx context.Context, salonID string) ([]domain.Service, error) {
	const q = `
SELECT id, salon_id, name, price_kzt, duration_min, is_active, created_at
FROM services
WHERE salon_id = $1
ORDER BY name
`
	rows, err := s.db.QueryContext(ctx, q, salonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Service{}
	for rows.Next() {
		var sv domain.Service
		if err := rows.Scan(&sv.ID, &sv.SalonID, &sv.Name, &sv.PriceKZT, &sv.DurationMin, &sv.IsActive, &sv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sv)
	}
	return out, rows.Err()
}

func (s *ServiceStore) GetByID(ctx context.Context, id string) (domain.Service, bool, error) {
	const q = `
SELECT id, salon_id, name, price_kzt, duration_min, is_active, created_at
FROM services WHERE id = $1
`
	var sv domain.Service
	err := s.db.QueryRowContext(ctx, q, id).Scan(
		&sv.ID, &sv.SalonID, &sv.Name, &sv.PriceKZT, &sv.DurationMin, &sv.IsActive, &sv.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Service{}, false, nil
	}
	if err != nil {
		return domain.Service{}, false, err
	}
	return sv, true, nil
}

func (s *ServiceStore) CreateForSalon(ctx context.Context, sv domain.Service) (domain.Service, error) {
	const q = `
INSERT INTO services (id, salon_id, name, price_kzt, duration_min, is_active)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING created_at
`
	err := s.db.QueryRowContext(ctx, q,
		sv.ID, sv.SalonID, sv.Name, sv.PriceKZT, sv.DurationMin, sv.IsActive,
	).Scan(&sv.CreatedAt)
	if err != nil {
		return domain.Service{}, mapPgError(err)
	}
	return sv, nil
}

func (s *ServiceStore) Update(ctx context.Context, sv domain.Service) (domain.Service, error) {
	const q = `
UPDATE services
SET name=$2, price_kzt=$3, duration_min=$4, is_active=$5
WHERE id=$1
RETURNING salon_id, created_at
`
	err := s.db.QueryRowContext(ctx, q,
		sv.ID, sv.Name, sv.PriceKZT, sv.DurationMin, sv.IsActive,
	).Scan(&sv.SalonID, &sv.CreatedAt)
	if err == sql.ErrNoRows {
		return domain.Service{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Service{}, mapPgError(err)
	}
	return sv, nil
}

func (s *ServiceStore) Disable(ctx context.Context, id string) error {
	const q = `UPDATE services SET is_active=false WHERE id=$1`
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
