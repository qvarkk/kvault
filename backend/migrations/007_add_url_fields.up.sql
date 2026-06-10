ALTER TABLE items
  ADD COLUMN source_url TEXT,
  ADD COLUMN url_metadata TEXT,
  ADD COLUMN extracted_content TEXT;

CREATE OR REPLACE FUNCTION update_search_vector_items()
RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(NEW.content, '')), 'B') ||
    setweight(to_tsvector('simple', coalesce(NEW.extracted_content, '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS items_search_vector_trigger ON items;

CREATE TRIGGER items_search_vector_trigger
BEFORE INSERT OR UPDATE OF title, content, extracted_content
ON items
FOR EACH ROW
EXECUTE FUNCTION update_search_vector_items();
