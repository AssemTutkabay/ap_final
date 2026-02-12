package postgres

import (
	"errors"

	"ap_final/internal/store"

	"github.com/jackc/pgx/v5/pgconn"
)

func mapPgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23P01 = exclusion_violation (overlap booking)
		// 23505 = unique_violation (email/phone, review booking_id, etc)
		if pgErr.Code == "23P01" || pgErr.Code == "23505" {
			return store.ErrConflict
		}
	}
	return err
}
