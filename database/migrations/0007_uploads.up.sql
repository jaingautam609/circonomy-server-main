CREATE TYPE upload_type AS ENUM ('contact','profile','certificate', 'misc');

CREATE TABLE IF NOT EXISTS uploads(
      id uuid PRIMARY KEY,
      path text NOT NULL,
      type upload_type NOT NULL,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      archived_at TIMESTAMP WITH TIME ZONE
);
