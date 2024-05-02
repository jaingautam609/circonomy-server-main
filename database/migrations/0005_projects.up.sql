CREATE TYPE status AS ENUM ('active', 'sold out', 'upcoming');
CREATE TYPE image_quality AS ENUM ('high quality', 'low quality');

CREATE TABLE IF NOT EXISTS certificates (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name TEXT NOT NULL,
     image_url TEXT NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
     archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS contacts (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name TEXT NOT NULL,
     image_url TEXT,
     description TEXT,
     email TEXT NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
     archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS projects (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name TEXT NOT NULL,
     project_time TIMESTAMP WITH TIME ZONE,
     capacity TEXT NOT NULL,
     image_urls TEXT[],
     location TEXT,
     total_cost TEXT,
     rate TEXT,
     method image_quality,
     description TEXT,
     certificates_ids UUID[],
     contacts_ids UUID[],
     project_status status,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
     archived_at TIMESTAMP WITH TIME ZONE
);