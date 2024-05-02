ALTER TABLE IF EXISTS projects_bought DROP COLUMN credits;
ALTER TABLE IF EXISTS projects_bought ADD COLUMN credits FLOAT;