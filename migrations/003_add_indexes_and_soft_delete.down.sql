DROP INDEX IF EXISTS items_user_id;
DROP INDEX IF EXISTS items_type;
DROP INDEX IF EXISTS items_created_at;
DROP INDEX IF EXISTS items_updated_at;
DROP INDEX IF EXISTS items_deleted_at;

DROP INDEX IF EXISTS files_user_id;
DROP INDEX IF EXISTS files_status;
DROP INDEX IF EXISTS files_created_at;
DROP INDEX IF EXISTS files_deleted_at;

ALTER TABLE items DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE files DROP COLUMN IF EXISTS deleted_at;