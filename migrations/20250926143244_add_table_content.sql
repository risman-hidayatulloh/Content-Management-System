-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS content_types (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        varchar(150) NOT NULL UNIQUE,
    slug        varchar(150) NOT NULL UNIQUE,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS field_defs (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type_id  uuid NOT NULL REFERENCES content_types(id) ON DELETE CASCADE,
    name             varchar(150) NOT NULL,
    key              varchar(150) NOT NULL,
    type             varchar(50)  NOT NULL, -- string|markdown|number|boolean|datetime|relation|json
    required         boolean NOT NULL DEFAULT false,
    is_unique        boolean NOT NULL DEFAULT false,
    is_relation      boolean NOT NULL DEFAULT false,
    relation_to      varchar(150),
    config           jsonb,
    "order"          int NOT NULL DEFAULT 0,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT field_defs_unique_per_ct UNIQUE (content_type_id, key)
);

CREATE TABLE IF NOT EXISTS entries (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type_id  uuid NOT NULL REFERENCES content_types(id) ON DELETE CASCADE,
    slug             text,                                   -- nullable
    status           varchar(16) NOT NULL DEFAULT 'draft',   -- draft|published
    scheduled_at     timestamptz,
    published_at     timestamptz,
    data             jsonb NOT NULL DEFAULT '{}'::jsonb,
    version          int NOT NULL DEFAULT 1,
    created_by       varchar(64),
    updated_by       varchar(64),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);

CREATE INDEX IF NOT EXISTS idx_entries_ctid          ON entries (content_type_id);
CREATE INDEX IF NOT EXISTS idx_entries_status        ON entries (status);
CREATE INDEX IF NOT EXISTS idx_entries_scheduled_at  ON entries (scheduled_at);
CREATE INDEX IF NOT EXISTS idx_entries_published_at  ON entries (published_at);
CREATE INDEX IF NOT EXISTS idx_entries_data_gin      ON entries USING GIN (data);

ALTER TABLE entries
  ADD CONSTRAINT uq_entries_slug_active UNIQUE (content_type_id, slug, deleted_at);

CREATE TABLE IF NOT EXISTS entry_versions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id   uuid NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    version    int  NOT NULL,
    data       jsonb NOT NULL,
    summary    text,
    actor_id   varchar(64),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT entry_versions_unique_ver UNIQUE (entry_id, version)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id   varchar(64),
    action     varchar(64) NOT NULL,
    target     varchar(128) NOT NULL,
    meta       jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_ct_updated_at
BEFORE UPDATE ON content_types
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_fd_updated_at
BEFORE UPDATE ON field_defs
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_entries_updated_at
BEFORE UPDATE ON entries
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DO $$
DECLARE
  v_ct_id uuid;
  v_entry_id uuid;
BEGIN
  INSERT INTO content_types (id, name, slug)
  VALUES (gen_random_uuid(), 'Article', 'article')
  ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
  RETURNING id INTO v_ct_id;

  IF v_ct_id IS NULL THEN
    SELECT id INTO v_ct_id FROM content_types WHERE slug = 'article';
  END IF;

  INSERT INTO field_defs (id, content_type_id, name, key, type, required, is_unique, is_relation, relation_to, "order")
  VALUES
    (gen_random_uuid(), v_ct_id, 'Title', 'title', 'string',   true,  true,  false, NULL, 1),
    (gen_random_uuid(), v_ct_id, 'Body',  'body',  'markdown', false, false, false, NULL, 2)
  ON CONFLICT (content_type_id, key) DO NOTHING;

  INSERT INTO entries (id, content_type_id, slug, status, data, version, created_by, updated_by, deleted_at)
  VALUES (
    gen_random_uuid(),
    v_ct_id,
    'hello-world',
    'draft',
    jsonb_build_object('title','Hello World','body','# Hello World\nThis is the first post.'),
    1,
    'system','system',
    NULL
  )
  ON CONFLICT (content_type_id, slug, deleted_at) DO NOTHING
  RETURNING id INTO v_entry_id;

  IF v_entry_id IS NULL THEN
    SELECT id INTO v_entry_id
    FROM entries
    WHERE content_type_id = v_ct_id
      AND slug = 'hello-world'
      AND deleted_at IS NULL;
  END IF;

  INSERT INTO entry_versions (id, entry_id, version, data, summary, actor_id)
  VALUES (
    gen_random_uuid(),
    v_entry_id,
    1,
    jsonb_build_object('title','Hello World','body','# Hello World\nThis is the first post.'),
    'initial creation',
    'system'
  )
  ON CONFLICT (entry_id, version) DO NOTHING;
END$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_entries_updated_at ON entries;
DROP TRIGGER IF EXISTS trg_fd_updated_at ON field_defs;
DROP TRIGGER IF EXISTS trg_ct_updated_at ON content_types;
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS entry_versions;
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS field_defs;
DROP TABLE IF EXISTS content_types;
-- +goose StatementEnd
