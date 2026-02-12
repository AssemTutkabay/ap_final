-- +goose Up
-- +goose StatementBegin

-- USERS (один bcrypt hash для демо, потом заменишь)
INSERT INTO users (id, role, full_name, email, password_hash)
VALUES
    ('u_platform', 'platform_admin', 'Platform Admin', 'platform@demo.local',
     '$2a$10$z7hYcJm8Fv7Qj1m2X9vVdO1m0o3GZcHqNwQw3yKqjQJQp3kX8Q9o2'),
    ('u_owner', 'salon_admin', 'Salon Owner', 'owner@demo.local',
     '$2a$10$z7hYcJm8Fv7Qj1m2X9vVdO1m0o3GZcHqNwQw3yKqjQJQp3kX8Q9o2'),
    ('u_client', 'client', 'Client User', 'client@demo.local',
     '$2a$10$z7hYcJm8Fv7Qj1m2X9vVdO1m0o3GZcHqNwQw3yKqjQJQp3kX8Q9o2'),
    ('u_master1', 'master', 'Master User', 'master1@demo.local',
     '$2a$10$z7hYcJm8Fv7Qj1m2X9vVdO1m0o3GZcHqNwQw3yKqjQJQp3kX8Q9o2')
    ON CONFLICT DO NOTHING;

-- SALONS
INSERT INTO salons (id, name, city, marketplace_enabled)
VALUES
    ('s1', 'Bloom Studio', 'Almaty', true),
    ('s2', 'Nova Beauty', 'Astana', true)
    ON CONFLICT DO NOTHING;

-- SALON ADMINS
INSERT INTO salon_admins (salon_id, user_id)
VALUES
    ('s1', 'u_owner'),
    ('s2', 'u_owner')
    ON CONFLICT DO NOTHING;

-- MASTERS (m1 привязан к u_master1)
INSERT INTO masters (id, salon_id, user_id, full_name, is_active)
VALUES
    ('m1', 's1', 'u_master1', 'Aruzhan K.', true),
    ('m2', 's1', NULL,        'Dana S.', true),
    ('m3', 's2', NULL,        'Madina T.', true)
    ON CONFLICT DO NOTHING;

-- SERVICES
INSERT INTO services (id, salon_id, name, price_kzt, duration_min, is_active)
VALUES
    ('sv1', 's1', 'Haircut', 7000, 60, true),
    ('sv2', 's1', 'Manicure', 9000, 90, true),
    ('sv3', 's2', 'Brow shaping', 5000, 30, true)
    ON CONFLICT DO NOTHING;

-- MASTER_SERVICES
INSERT INTO master_services (master_id, service_id)
VALUES
    ('m1', 'sv1'),
    ('m2', 'sv2'),
    ('m3', 'sv3')
    ON CONFLICT DO NOTHING;

-- пример completed booking
INSERT INTO bookings (salon_id, client_id, master_id, service_id, start_at, end_at, status)
VALUES
    ('s1', 'u_client', 'm1', 'sv1',
     now() - interval '2 days',
     now() - interval '2 days' + interval '60 minutes',
     'completed')
    ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM audit_log WHERE actor_user_id IN ('u_platform','u_owner','u_client','u_master1');

DELETE FROM reviews WHERE client_id = 'u_client';
DELETE FROM bookings WHERE client_id = 'u_client';

DELETE FROM master_services WHERE master_id IN ('m1','m2','m3');
DELETE FROM services WHERE id IN ('sv1','sv2','sv3');
DELETE FROM masters WHERE id IN ('m1','m2','m3');

DELETE FROM salon_admins WHERE user_id = 'u_owner';
DELETE FROM salons WHERE id IN ('s1','s2');

DELETE FROM users WHERE id IN ('u_platform','u_owner','u_client','u_master1');

-- +goose StatementEnd