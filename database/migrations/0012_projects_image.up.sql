ALTER TABLE IF EXISTS clients DROP COLUMN IF EXISTS grayscale_path;

ALTER TABLE IF EXISTS projects DROP COLUMN IF EXISTS upload_ids;
ALTER TABLE IF EXISTS projects ADD COLUMN upload_id UUID REFERENCES uploads(id);
ALTER TYPE image_quality ADD VALUE 'project';