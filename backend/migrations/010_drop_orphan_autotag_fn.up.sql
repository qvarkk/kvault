-- Migration 005 dropped the insert-time auto-tag trigger and replaced
-- extract_item_tags with a 3-arg form, but because the signatures differ the old
-- 4-arg overload was never removed and lingered as dead code. Drop it. Auto-tagging
-- is intentionally manual-only now (POST /items/{id}/autotag).
DROP FUNCTION IF EXISTS extract_item_tags(uuid, uuid, text, tsvector);
