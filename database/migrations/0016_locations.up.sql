ALTER TABLE IF EXISTS users RENAME COLUMN location TO address;
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS lat FLOAT;
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS long FLOAT;

ALTER TABLE IF EXISTS projects DROP COLUMN IF EXISTS state;
ALTER TABLE IF EXISTS projects DROP COLUMN IF EXISTS country;

ALTER TABLE IF EXISTS projects ADD COLUMN IF NOT EXISTS lat FLOAT;
ALTER TABLE IF EXISTS projects ADD COLUMN IF NOT EXISTS long FLOAT;