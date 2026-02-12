package postgres

import (
	"context"
	"database/sql"
)

type SalonAdminStore struct{ db *sql.DB }

func NewSalonAdminStore(db *sql.DB) *SalonAdminStore { return &SalonAdminStore{db: db} }

func (s *SalonAdminStore) IsAdminOfSalon(ctx context.Context, userID, salonID string) (bool, error) {
	const q = `
SELECT 1
FROM salon_admins
WHERE user_id = $1 AND salon_id = $2
`
	var one int
	err := s.db.QueryRowContext(ctx, q, userID, salonID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SalonAdminStore) ListSalonIDsForAdmin(ctx context.Context, userID string) ([]string, error) {
	const q = `
SELECT salon_id
FROM salon_admins
WHERE user_id = $1
ORDER BY salon_id
`
	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
