ALTER TABLE items ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE files ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS items_user_id ON items(user_id);
CREATE INDEX IF NOT EXISTS items_type ON items(type);
CREATE INDEX IF NOT EXISTS items_created_at ON items(created_at DESC);
CREATE INDEX IF NOT EXISTS items_updated_at ON items(updated_at DESC);
CREATE INDEX IF NOT EXISTS items_deleted_at ON items(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS files_status ON files(status);
CREATE INDEX IF NOT EXISTS files_created_at ON files(created_at DESC);
CREATE INDEX IF NOT EXISTS files_deleted_at ON files(deleted_at) WHERE deleted_at IS NULL;