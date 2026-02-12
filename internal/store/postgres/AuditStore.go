package postgres

import (
	"context"
	"database/sql"

	"ap_final/internal/domain"
)

type AuditStore struct{ db *sql.DB }

func NewAuditStore(db *sql.DB) *AuditStore { return &AuditStore{db: db} }

func (s *AuditStore) Insert(ctx context.Context, ev domain.AuditLog) error {
	const q = `
INSERT INTO audit_log (actor_user_id, action, entity_type, entity_id, at, details)
VALUES ($1,$2,$3,$4,$5,$6)
`
	_, err := s.db.ExecContext(ctx, q,
		ev.ActorUserID,
		ev.Action,
		ev.EntityType,
		ev.EntityID,
		ev.At,
		ev.Details,
	)
	return mapPgError(err)
}
