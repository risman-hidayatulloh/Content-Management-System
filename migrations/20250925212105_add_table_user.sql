-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         varchar(255) UNIQUE NOT NULL,
    password_hash text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS roles (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       varchar(100) NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS permissions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    action     varchar(32) NOT NULL,
    resource   varchar(128) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT permissions_action_chk CHECK (action IN ('create','read','edit','publish','delete')),
    CONSTRAINT permissions_unique UNIQUE (action, resource)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id uuid NOT NULL,
    role_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_user_roles_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role
        FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       uuid NOT NULL,
    permission_id uuid NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role_permissions_role
        FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission
        FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_roles_updated_at
BEFORE UPDATE ON roles
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_permissions_updated_at
BEFORE UPDATE ON permissions
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =========================
-- SEED (idempotent)
-- =========================

-- role Admin
INSERT INTO roles (id, name)
VALUES (gen_random_uuid(), 'Admin')
ON CONFLICT (name) DO NOTHING;

-- permissions dasar
INSERT INTO permissions (id, action, resource)
VALUES
  (gen_random_uuid(),'create','content:*'),
  (gen_random_uuid(),'read','content:*'),
  (gen_random_uuid(),'edit','content:*'),
  (gen_random_uuid(),'publish','content:*'),
  (gen_random_uuid(),'delete','content:*'),
  (gen_random_uuid(),'create','media:*'),
  (gen_random_uuid(),'read','media:*')
ON CONFLICT (action, resource) DO NOTHING;

-- admin user + relasi (email & password bisa kamu ubah di sini)
DO $$
DECLARE v_user_id uuid;
BEGIN
  INSERT INTO users (id, email, password_hash, created_at, updated_at)
  VALUES (gen_random_uuid(), 'admin@example.com',
          crypt('admin123', gen_salt('bf')), now(), now())
  ON CONFLICT (email) DO UPDATE
    SET password_hash = EXCLUDED.password_hash,
        updated_at = now();

  SELECT id INTO v_user_id FROM users WHERE email = 'admin@example.com';

  -- link user -> Admin role
  INSERT INTO user_roles (user_id, role_id)
  SELECT v_user_id, r.id
  FROM roles r
  WHERE r.name = 'Admin'
  ON CONFLICT (user_id, role_id) DO NOTHING;

  -- link Admin role -> semua permissions
  INSERT INTO role_permissions (role_id, permission_id)
  SELECT r.id, p.id
  FROM roles r
  JOIN permissions p ON TRUE
  WHERE r.name = 'Admin'
  ON CONFLICT (role_id, permission_id) DO NOTHING;
END$$;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_permissions_updated_at ON permissions;
DROP TRIGGER IF EXISTS trg_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;

-- DROP EXTENSION IF EXISTS pgcrypto;
-- +goose StatementEnd
