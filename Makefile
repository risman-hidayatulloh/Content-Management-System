SRVDIR := server
MIGDIR := ../migrations

dev:
	@echo "Starting Application (dev hot-reload)..."
	docker compose --profile dev up

prod:
	@echo "Starting Application (prod)..."
	docker compose --profile prod up --build

rebuild:
	docker compose --profile dev build server

down:
	docker compose down

tidy:
	@ cd $(SRVDIR) && go mod tidy

migrate-create:
	@ echo "---Creating migration files---"
	@ cd $(SRVDIR) && mkdir -p $(MIGDIR) && go run ./cmd/migrate/main.go -dir $(MIGDIR) create $(NAME) sql

migrate-up:
	@ cd $(SRVDIR) && go run ./cmd/migrate/main.go -dir $(MIGDIR) up

migrate-down:
	@ cd $(SRVDIR) && go run ./cmd/migrate/main.go -dir $(MIGDIR) down

migrate-down-to:
	@ cd $(SRVDIR) && go run ./cmd/migrate/main.go -dir $(MIGDIR) down-to $(VERSION)

migrate-force:
	@ cd $(SRVDIR) && go run ./cmd/migrate/main.go -dir $(MIGDIR) force $(VERSION)

migrate-version:
	@ cd $(SRVDIR) && go run ./cmd/migrate/main.go -dir $(MIGDIR) version

seed:
	@ cat <<'SQL' | docker compose exec -T db psql -U postgres -d cms -v ON_ERROR_STOP=1;

	-- roles
	INSERT INTO roles (id, name)
	VALUES (gen_random_uuid(), 'Admin')
	ON CONFLICT (name) DO NOTHING;

	-- permissions (idempotent)
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

	-- admin user (update hash jika sudah ada)
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
	  SELECT v_user_id, r.id FROM roles r WHERE r.name = 'Admin'
	  ON CONFLICT (user_id, role_id) DO NOTHING;

	  -- link Admin role -> all permissions
	  INSERT INTO role_permissions (role_id, permission_id)
	  SELECT r.id, p.id
	  FROM roles r CROSS JOIN permissions p
	  WHERE r.name = 'Admin'
	  ON CONFLICT (role_id, permission_id) DO NOTHING;
	END$$;
	SQL
	@ echo "ok"
