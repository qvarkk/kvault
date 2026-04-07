DROP TRIGGER IF EXISTS items_search_vector_trigger ON items;
DROP TRIGGER IF EXISTS files_search_vector_trigger ON files;

DROP FUNCTION IF EXISTS update_search_vector_items();
DROP FUNCTION IF EXISTS update_search_vector_files();