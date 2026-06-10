CREATE TYPE tag_source AS ENUM ('auto', 'manual');
ALTER TABLE item_tags ADD COLUMN source tag_source NOT NULL DEFAULT 'auto';

CREATE TYPE stopword_source AS ENUM ('user', 'default');
CREATE TABLE IF NOT EXISTS stopwords (
  word TEXT NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id),
  source stopword_source NOT NULL,
  is_enabled BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, word)
);
CREATE INDEX idx_stopwords_user_id ON stopwords(user_id);


CREATE TABLE IF NOT EXISTS stopwords_default (
  word TEXT PRIMARY KEY
);

-- AI GENERATED, MIGHT BE A SUBJECT TO CHANGE
INSERT INTO stopwords_default (word) VALUES
  -- english articles, pronouns, prepositions
  ('the'), ('and'), ('that'), ('this'), ('with'), ('from'), ('have'), ('will'),
  ('been'), ('were'), ('they'), ('them'), ('their'), ('there'), ('these'),
  ('those'), ('what'), ('which'), ('when'), ('where'), ('while'), ('then'),
  ('than'), ('such'), ('some'), ('more'), ('most'), ('also'), ('just'),
  ('very'), ('much'), ('many'), ('even'), ('only'), ('back'), ('over'),
  ('after'), ('before'), ('about'), ('above'), ('below'), ('under'), ('into'),
  ('onto'), ('upon'), ('through'), ('between'), ('around'), ('against'),
  ('without'), ('within'), ('during'), ('because'), ('though'), ('although'),
  ('however'), ('therefore'), ('moreover'), ('furthermore'), ('nevertheless'),
  ('both'), ('either'), ('neither'), ('each'), ('every'), ('other'), ('another'),
  ('same'), ('different'), ('still'), ('already'), ('always'), ('never'),
  ('often'), ('again'), ('here'), ('there'), ('well'), ('like'), ('make'),
  ('made'), ('take'), ('taken'), ('come'), ('came'), ('gone'), ('give'),
  ('given'), ('know'), ('known'), ('think'), ('thought'), ('look'), ('looks'),
  ('need'), ('needs'), ('want'), ('wants'), ('used'), ('uses'), ('said'),
  ('says'), ('does'), ('doing'), ('done'), ('gets'), ('getting'), ('got'),
  ('let'), ('lets'), ('help'), ('helps'), ('keep'), ('keeps'), ('show'),
  ('shows'), ('find'), ('found'), ('call'), ('calls'), ('seem'), ('seems'),
  ('feel'), ('feels'), ('become'), ('became'), ('leave'), ('left'), ('right'),
  ('might'), ('could'), ('would'), ('should'), ('shall'), ('must'), ('really'),
  ('quite'), ('rather'), ('almost'), ('enough'), ('else'), ('once'), ('since'),
  ('until'), ('unless'), ('whether'), ('instead'), ('along'), ('away'),
  ('down'), ('your'), ('mine'), ('ours'), ('them'), ('itself'), ('himself'),
  ('herself'), ('myself'), ('yourself'), ('ourselves'), ('themselves'),

  -- russian pronouns, particles, prepositions, conjunctions
  ('это'), ('как'), ('все'), ('так'), ('что'), ('при'), ('или'), ('если'),
  ('когда'), ('чтобы'), ('было'), ('быть'), ('есть'), ('нет'), ('уже'),
  ('еще'), ('вот'), ('даже'), ('тоже'), ('также'), ('либо'), ('хотя'),
  ('пока'), ('после'), ('перед'), ('между'), ('через'), ('около'), ('вдоль'),
  ('против'), ('кроме'), ('вместо'), ('среди'), ('ради'), ('ввиду'), ('вследствие'),
  ('благодаря'), ('несмотря'), ('потому'), ('поэтому'), ('однако'), ('зато'),
  ('притом'), ('причем'), ('итак'), ('значит'), ('следовательно'), ('например'),
  ('кстати'), ('впрочем'), ('наконец'), ('сначала'), ('потом'), ('затем'),
  ('сразу'), ('снова'), ('опять'), ('просто'), ('только'), ('лишь'), ('именно'),
  ('вдруг'), ('почти'), ('совсем'), ('очень'), ('весьма'), ('довольно'),
  ('quite'), ('около'), ('будто'), ('словно'), ('якобы'), ('якобы'),
  ('мной'), ('тебя'), ('тебе'), ('тобой'), ('него'), ('ней'), ('них'),
  ('ним'), ('ними'), ('себя'), ('себе'), ('собой'), ('свой'), ('своя'),
  ('свое'), ('свои'), ('этот'), ('эта'), ('эти'), ('того'), ('той'),
  ('тех'), ('тому'), ('тем'), ('том'), ('той'), ('такой'), ('такая'),
  ('такие'), ('такого'), ('такой'), ('каждый'), ('каждая'), ('каждое'),
  ('любой'), ('любая'), ('любое'), ('весь'), ('вся'), ('всего'), ('всей'),
  ('всем'), ('всею'), ('один'), ('одна'), ('одно'), ('одни'), ('другой'),
  ('другая'), ('другое'), ('другие'), ('самый'), ('самая'), ('самое'),
  ('можно'), ('нужно'), ('надо'), ('нельзя'), ('видно'), ('известно'),
  ('понятно'), ('ясно'), ('трудно'), ('легко'), ('важно'), ('интересно'),
  ('хорошо'), ('плохо'), ('много'), ('мало'), ('немного'), ('несколько'),
  ('сколько'), ('столько'), ('больше'), ('меньше'), ('лучше'), ('хуже'),
  ('раньше'), ('позже'), ('быстро'), ('медленно'), ('долго'), ('скоро')
ON CONFLICT DO NOTHING;


CREATE OR REPLACE FUNCTION active_stopwords(p_user_id UUID)
RETURNS TABLE(word TEXT, source stopword_source, is_enabled BOOL, updated_at TIMESTAMPTZ) AS $$
  SELECT 
    sd.word,
    'default'::stopword_source AS source,
    CASE WHEN s.word IS NOT NULL THEN s.is_enabled ELSE true END AS is_enabled,
    COALESCE(s.updated_at, '0001-01-01 00:00:00+00'::TIMESTAMPTZ) AS updated_at
  FROM stopwords_default sd
  LEFT JOIN stopwords s ON s.word = sd.word AND s.user_id = p_user_id

  UNION ALL

  SELECT
    s.word,
    'user'::stopword_source AS source,
    s.is_enabled,
    s.updated_at
  FROM stopwords s
  WHERE s.user_id = p_user_id
    AND s.word NOT IN (SELECT word FROM stopwords_default);
$$ LANGUAGE sql STABLE;


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
      '^[a-zA-Zа-яА-ЯёЁ\u00C0-\u024F]+$',
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


CREATE OR REPLACE FUNCTION trigger_auto_tag_item()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM extract_item_tags(NEW.id, NEW.user_id, NEW.content, NEW.search_vector);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER auto_tag_item_update
  AFTER INSERT ON items
  FOR EACH ROW
  EXECUTE FUNCTION trigger_auto_tag_item();