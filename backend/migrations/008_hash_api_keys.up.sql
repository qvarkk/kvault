-- API keys are now stored as a SHA-256 hash, never as plaintext.
-- No data backfill: there are no real users yet, so existing rows (if any in dev)
-- simply have their column renamed; their old plaintext values become unusable
-- and the affected user must rotate/re-login to obtain a working key.
ALTER TABLE users RENAME COLUMN api_key TO api_key_hash;
