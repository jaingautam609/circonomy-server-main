ALTER TYPE business_size RENAME TO business_size_old;
CREATE TYPE business_size AS ENUM ('0-50', '51-100', '100+');
ALTER TABLE users ALTER COLUMN size TYPE business_size USING size::text::business_size;


