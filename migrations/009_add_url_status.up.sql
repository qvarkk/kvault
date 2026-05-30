-- Track the lifecycle of a url-type item's background fetch so the UI can tell
-- pending / ready / error apart (previously indistinguishable). Nullable; only
-- url-type items use it.
ALTER TABLE items ADD COLUMN IF NOT EXISTS url_status TEXT;
