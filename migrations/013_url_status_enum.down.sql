ALTER TABLE items
  ALTER COLUMN url_status TYPE TEXT USING url_status::text;

DROP TYPE url_status;
