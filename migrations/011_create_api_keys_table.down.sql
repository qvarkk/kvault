-- Lossy reverse: restores a single api_key_hash per user (the most recently
-- created key); any additional per-device keys are discarded.
ALTER TABLE users ADD COLUMN api_key_hash TEXT;

UPDATE users u
SET api_key_hash = k.key_hash
FROM (
  SELECT DISTINCT ON (user_id) user_id, key_hash
  FROM api_keys
  ORDER BY user_id, created_at DESC
) k
WHERE k.user_id = u.id;

ALTER TABLE users ALTER COLUMN api_key_hash SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_api_key_hash_key UNIQUE (api_key_hash);

DROP TABLE api_keys;
