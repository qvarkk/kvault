DROP TRIGGER IF EXISTS auto_tag_item_update ON ITEMS;
DROP FUNCTION IF EXISTS trigger_auto_tag_item();
DROP FUNCTION IF EXISTS extract_item_tags();
DROP FUNCTION IF EXISTS active_stopwords();

DROP TABLE IF EXISTS stopwords_default;

DROP TABLE IF EXISTS stopwords;
DROP TYPE IF EXISTS stopword_source;
DROP INDEX IF EXISTS idx_stopwords_user_id;

ALTER TABLE item_tags DROP COLUMN IF EXISTS source;
DROP TYPE IF EXISTS tag_source;