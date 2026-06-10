DROP TYPE IF EXISTS url_status;
CREATE TYPE url_status AS ENUM ('pending', 'ready', 'error');

-- Column stays nullable: only url-type items use it. The USING cast converts the
-- existing pending/ready/error text values in place.
ALTER TABLE items
  ALTER COLUMN url_status TYPE url_status USING url_status::url_status;
