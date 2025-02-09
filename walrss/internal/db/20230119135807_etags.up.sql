ALTER TABLE feeds ADD COLUMN last_etag VARCHAR
--bun:split
ALTER TABLE feeds ADD COLUMN cached_content VARCHAR