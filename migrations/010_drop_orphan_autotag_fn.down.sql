-- Recreate the 4-arg overload exactly as migration 004 defined it (reversibility).
CREATE OR REPLACE FUNCTION extract_item_tags(item_id UUID, item_user_id UUID, content TEXT, search_vector tsvector)
RETURNS VOID AS $$
DECLARE
  tag_word TEXT;
  tag_id   UUID;
BEGIN
  IF length(coalesce(content, '')) < 50 THEN
    RETURN;
  END IF;

  FOR tag_word IN
    EXECUTE format(
      'SELECT word FROM ts_stat(%L)
        WHERE length(word) > 3
          AND word ~ %L
          AND word NOT IN (SELECT word FROM active_stopwords(%L::uuid))
        ORDER BY nentry DESC
        LIMIT 3',
      'SELECT search_vector FROM items WHERE id = ''' || item_id || '''',
      '^[a-zA-Zа-яА-ЯёЁÀ-ɏ]+$',
      item_user_id
    )
  LOOP
    INSERT INTO tags (user_id, name)
    VALUES (item_user_id, tag_word)
    ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO tag_id;

    INSERT INTO item_tags (item_id, tag_id, source)
    VALUES (item_id, tag_id, 'auto')
    ON CONFLICT DO NOTHING;
  END LOOP;
END;
$$ LANGUAGE plpgsql;
