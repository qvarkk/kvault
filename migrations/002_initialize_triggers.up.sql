CREATE OR REPLACE FUNCTION update_search_vector_items()
RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(NEW.content, '')), 'B');

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_search_vector_files()
RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('simple', coalesce(NEW.extracted_content, ''));

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER items_search_vector_trigger
BEFORE INSERT OR UPDATE OF title, content
ON items
FOR EACH ROW
EXECUTE FUNCTION update_search_vector_items();

CREATE TRIGGER files_search_vector_trigger
BEFORE INSERT OR UPDATE OF original_name, extracted_content
ON files
FOR EACH ROW
EXECUTE FUNCTION update_search_vector_files();