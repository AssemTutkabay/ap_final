-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS btree_gist;

-- users
CREATE TABLE IF NOT EXISTS users (
                                     id            text PRIMARY KEY,
                                     role          text NOT NULL CHECK (role IN ('client','salon_admin','master','platform_admin')),
    full_name     text,
    email         text,
    phone         text,
    password_hash text NOT NULL,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    CHECK (email IS NOT NULL OR phone IS NOT NULL)
    );

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users(email) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_phone ON users(phone) WHERE phone IS NOT NULL;

-- salons
CREATE TABLE IF NOT EXISTS salons (
                                      id                  text PRIMARY KEY,
                                      name                text NOT NULL,
                                      city                text NOT NULL,
                                      marketplace_enabled boolean NOT NULL DEFAULT true,
                                      created_at          timestamptz NOT NULL DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_salons_city ON salons (city);

-- salon_admins (mapping)
CREATE TABLE IF NOT EXISTS salon_admins (
                                            salon_id   text NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    user_id    text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (salon_id, user_id)
    );

CREATE INDEX IF NOT EXISTS idx_salon_admins_user ON salon_admins(user_id);
CREATE INDEX IF NOT EXISTS idx_salon_admins_salon ON salon_admins(salon_id);

-- masters (WITH user_id binding)
CREATE TABLE IF NOT EXISTS masters (
                                       id            text PRIMARY KEY,
                                       salon_id       text NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    user_id        text REFERENCES users(id) ON DELETE SET NULL,
    full_name      text NOT NULL,
    bio            text,
    portfolio_url  text,
    is_active      boolean NOT NULL DEFAULT true,
    created_at     timestamptz NOT NULL DEFAULT now()
    );

CREATE UNIQUE INDEX IF NOT EXISTS ux_masters_user_id ON masters(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_masters_salon ON masters(salon_id);

-- services
CREATE TABLE IF NOT EXISTS services (
                                        id           text PRIMARY KEY,
                                        salon_id     text NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    name         text NOT NULL,
    price_kzt    int NOT NULL CHECK (price_kzt >= 0),
    duration_min int NOT NULL CHECK (duration_min > 0),
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_services_salon ON services(salon_id);

-- master_services (M2M)
CREATE TABLE IF NOT EXISTS master_services (
                                               master_id  text NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    service_id text NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (master_id, service_id)
    );

-- bookings
CREATE TABLE IF NOT EXISTS bookings (
                                        id             bigserial PRIMARY KEY,
                                        salon_id       text NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    client_id      text NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    master_id      text NOT NULL REFERENCES masters(id) ON DELETE RESTRICT,
    service_id     text NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    start_at       timestamptz NOT NULL,
    end_at         timestamptz NOT NULL,
    status         text NOT NULL CHECK (status IN ('created','confirmed','cancelled_by_client','cancelled_by_salon','completed')),
    cancel_reason  text,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CHECK (start_at < end_at)
    );

CREATE INDEX IF NOT EXISTS idx_bookings_client ON bookings(client_id);
CREATE INDEX IF NOT EXISTS idx_bookings_master ON bookings(master_id, start_at);

-- no overlap for active bookings
ALTER TABLE bookings
    ADD CONSTRAINT bookings_no_overlap_active
    EXCLUDE USING gist (
    master_id WITH =,
    tstzrange(start_at, end_at, '[)') WITH &&
  )
  WHERE (status IN ('created','confirmed'));

-- reviews (verified)
CREATE TABLE IF NOT EXISTS reviews (
                                       id         bigserial PRIMARY KEY,
                                       booking_id bigint NOT NULL UNIQUE REFERENCES bookings(id) ON DELETE CASCADE,
    client_id  text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    salon_id   text NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    master_id  text NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    rating     int NOT NULL CHECK (rating BETWEEN 1 AND 5),
    text       text,
    created_at timestamptz NOT NULL DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_reviews_salon ON reviews(salon_id);

-- audit_log
CREATE TABLE IF NOT EXISTS audit_log (
                                         id           bigserial PRIMARY KEY,
                                         actor_user_id text REFERENCES users(id) ON DELETE SET NULL,
    action       text NOT NULL,
    entity_type  text NOT NULL,
    entity_id    text,
    at           timestamptz NOT NULL DEFAULT now(),
    details      jsonb NOT NULL DEFAULT '{}'::jsonb
    );

CREATE INDEX IF NOT EXISTS idx_audit_at ON audit_log(at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS reviews;
ALTER TABLE IF EXISTS bookings DROP CONSTRAINT IF EXISTS bookings_no_overlap_active;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS master_services;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS masters;
DROP TABLE IF EXISTS salon_admins;
DROP TABLE IF EXISTS salons;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd