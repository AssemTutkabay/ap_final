package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"ap_final/internal/domain"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore { return &UserStore{db: db} }

func (s *UserStore) Create(ctx context.Context, u domain.User) (domain.User, error) {
	const q = `
INSERT INTO users (id, role, full_name, email, phone, password_hash, is_active)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING created_at
`
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Phone = strings.TrimSpace(u.Phone)

	var createdAt timeNull
	err := s.db.QueryRowContext(ctx, q,
		u.ID,
		string(u.Role),
		nullIfEmpty(strings.TrimSpace(u.FullName)),
		nullIfEmpty(u.Email),
		nullIfEmpty(u.Phone),
		u.PasswordHash,
		u.IsActive,
	).Scan(&createdAt)
	if err != nil {
		return domain.User{}, mapPgError(err)
	}

	u.CreatedAt = createdAt.Time
	return u, nil
}

func (s *UserStore) GetByID(ctx context.Context, id string) (domain.User, bool, error) {
	const q = `
SELECT id, role, full_name, email, phone, password_hash, is_active, created_at
FROM users WHERE id = $1
`
	return s.getOne(ctx, q, id)
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (domain.User, bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return domain.User{}, false, nil
	}
	const q = `
SELECT id, role, full_name, email, phone, password_hash, is_active, created_at
FROM users WHERE email = $1
`
	return s.getOne(ctx, q, email)
}

func (s *UserStore) GetByPhone(ctx context.Context, phone string) (domain.User, bool, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return domain.User{}, false, nil
	}
	const q = `
SELECT id, role, full_name, email, phone, password_hash, is_active, created_at
FROM users WHERE phone = $1
`
	return s.getOne(ctx, q, phone)
}

func (s *UserStore) getOne(ctx context.Context, q string, arg any) (domain.User, bool, error) {
	var u domain.User
	var role string
	var fullName, email, phone sql.NullString
	var createdAt timeNull

	err := s.db.QueryRowContext(ctx, q, arg).Scan(
		&u.ID,
		&role,
		&fullName,
		&email,
		&phone,
		&u.PasswordHash,
		&u.IsActive,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, err
	}

	u.Role = domain.UserRole(role)
	if fullName.Valid {
		u.FullName = fullName.String
	}
	if email.Valid {
		u.Email = email.String
	}
	if phone.Valid {
		u.Phone = phone.String
	}
	u.CreatedAt = createdAt.Time

	return u, true, nil
}

// ---- helpers (local) ----

type timeNull struct {
	Time  time.Time
	Valid bool
}

func (t *timeNull) Scan(value any) error {
	var nt sql.NullTime
	if err := nt.Scan(value); err != nil {
		return err
	}
	t.Valid = nt.Valid
	t.Time = nt.Time
	return nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
