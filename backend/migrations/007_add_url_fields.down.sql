ALTER TABLE items
  DROP COLUMN IF EXISTS source_url,
  DROP COLUMN IF EXISTS url_metadata,
  DROP COLUMN IF EXISTS extracted_content;

CREATE OR REPLACE FUNCTION update_search_vector_items()
RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(NEW.content, '')), 'B');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS items_search_vector_trigger ON items;

CREATE TRIGGER items_search_vector_trigger
BEFORE INSERT OR UPDATE OF title, content
ON items
FOR EACH ROW
EXECUTE FUNCTION update_search_vector_items();
