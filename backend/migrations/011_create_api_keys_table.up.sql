CREATE TABLE IF NOT EXISTS api_keys (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  key_hash     TEXT NOT NULL UNIQUE,
  label        TEXT NOT NULL DEFAULT '',
  last_used_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

-- Backfill each user's existing single key so live sessions keep working after
-- the column is dropped. 30d matches the default AUTH_API_KEY_TTL.
INSERT INTO api_keys (user_id, key_hash, label, last_used_at, created_at, expires_at)
SELECT id, api_key_hash, 'Legacy', now(), now(), now() + interval '30 days'
FROM users;

ALTER TABLE users DROP COLUMN api_key_hash;
