package postgres

import (
	"context"
	"database/sql"
	"strings"

	"ap_final/internal/domain"
	"ap_final/internal/store"
)

type SalonStore struct{ db *sql.DB }

func NewSalonStore(db *sql.DB) *SalonStore { return &SalonStore{db: db} }

func (s *SalonStore) ListPublic(ctx context.Context, city string) ([]domain.Salon, error) {
	city = strings.TrimSpace(city)

	q := `
SELECT id, name, city, marketplace_enabled, created_at
FROM salons
WHERE marketplace_enabled = true
`
	args := []any{}
	if city != "" {
		q += " AND lower(city) = lower($1)"
		args = append(args, city)
	}
	q += " ORDER BY name"

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Salon{}
	for rows.Next() {
		var sal domain.Salon
		if err := rows.Scan(&sal.ID, &sal.Name, &sal.City, &sal.MarketplaceEnabled, &sal.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sal)
	}
	return out, rows.Err()
}

func (s *SalonStore) GetByID(ctx context.Context, id string) (domain.Salon, bool, error) {
	const q = `
SELECT id, name, city, marketplace_enabled, created_at
FROM salons WHERE id = $1
`
	var sal domain.Salon
	err := s.db.QueryRowContext(ctx, q, id).Scan(&sal.ID, &sal.Name, &sal.City, &sal.MarketplaceEnabled, &sal.CreatedAt)
	if err == sql.ErrNoRows {
		return domain.Salon{}, false, nil
	}
	if err != nil {
		return domain.Salon{}, false, err
	}
	return sal, true, nil
}

func (s *SalonStore) Create(ctx context.Context, sal domain.Salon) (domain.Salon, error) {
	const q = `
INSERT INTO salons (id, name, city, marketplace_enabled)
VALUES ($1,$2,$3,$4)
RETURNING created_at
`
	err := s.db.QueryRowContext(ctx, q, sal.ID, sal.Name, sal.City, sal.MarketplaceEnabled).Scan(&sal.CreatedAt)
	if err != nil {
		return domain.Salon{}, mapPgError(err)
	}
	return sal, nil
}

func (s *SalonStore) Update(ctx context.Context, sal domain.Salon) (domain.Salon, error) {
	const q = `
UPDATE salons
SET name=$2, city=$3, marketplace_enabled=$4
WHERE id=$1
RETURNING created_at
`
	err := s.db.QueryRowContext(ctx, q, sal.ID, sal.Name, sal.City, sal.MarketplaceEnabled).Scan(&sal.CreatedAt)
	if err == sql.ErrNoRows {
		return domain.Salon{}, store.ErrNotFound
	}
	if err != nil {
		return domain.Salon{}, mapPgError(err)
	}
	return sal, nil
}

func (s *SalonStore) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM salons WHERE id = $1`
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
