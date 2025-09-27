-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    filename    varchar(255) NOT NULL,
    mime        varchar(100) NOT NULL,
    size        bigint NOT NULL,
    width       int,
    height      int,
    variants    jsonb, -- bisa simpan info resize / thumbnail
    url         text NOT NULL,
    storage_key text NOT NULL,
    created_by  varchar(64) NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_media_created_by ON media (created_by);
CREATE INDEX IF NOT EXISTS idx_media_mime ON media (mime);

INSERT INTO media (id, filename, mime, size, width, height, variants, url, storage_key, created_by)
VALUES (
    gen_random_uuid(),
    'hello.png',
    'image/png',
    12345,
    800,
    600,
    jsonb_build_object('thumbnail','/media/hello-thumb.png'),
    '/media/hello.png',
    'uploads/2025/09/hello.png',
    'system'
)
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_created_by;
DROP INDEX IF EXISTS idx_media_mime;
DROP TABLE IF EXISTS media;
-- +goose StatementEnd
