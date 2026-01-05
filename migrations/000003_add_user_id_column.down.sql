ALTER TABLE shorten_urls
DROP COLUMN user_id,
DROP COLUMN IF EXISTS is_deleted;