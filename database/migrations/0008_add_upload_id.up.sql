ALTER TABLE IF EXISTS certificates DROP COLUMN image_url;
ALTER TABLE IF EXISTS certificates ADD COLUMN IF NOT EXISTS upload_id uuid references uploads(id);

ALTER TABLE IF EXISTS contacts DROP COLUMN image_url;
ALTER TABLE IF EXISTS contacts ADD COLUMN IF NOT EXISTS upload_id uuid references uploads(id);

ALTER TABLE IF EXISTS projects DROP COLUMN image_urls;
ALTER TABLE IF EXISTS projects ADD COLUMN IF NOT EXISTS upload_ids uuid[];

ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS upload_id uuid references uploads(id);
