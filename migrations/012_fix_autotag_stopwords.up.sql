-- active_stopwords() returns every stopword with an is_enabled flag, so the
-- autotag filter must check the flag — otherwise disabled stopwords are still
-- excluded and can never become tags.
CREATE OR REPLACE FUNCTION extract_item_tags(p_item_id UUID, p_user_id UUID, p_tag_count INT DEFAULT 3)
RETURNS VOID AS $$
DECLARE
  tag_word TEXT;
  tag_id   UUID;
BEGIN
  FOR tag_word IN
    EXECUTE format(
      'SELECT word FROM ts_stat(%L)
        WHERE length(word) > 3
          AND word ~ %L
          AND word NOT IN (SELECT word FROM active_stopwords(%L::uuid) WHERE is_enabled = true)
        ORDER BY nentry DESC
        LIMIT %s',
      'SELECT search_vector FROM items WHERE id = ''' || p_item_id || '''',
      '^[a-zA-Zа-яА-ЯёЁÀ-ɏ]+$',
      p_user_id,
      p_tag_count
    )
  LOOP
    INSERT INTO tags (user_id, name)
    VALUES (p_user_id, tag_word)
    ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO tag_id;

    INSERT INTO item_tags (item_id, tag_id, source)
    VALUES (p_item_id, tag_id, 'auto')
    ON CONFLICT DO NOTHING;
  END LOOP;
END;
$$ LANGUAGE plpgsql;
