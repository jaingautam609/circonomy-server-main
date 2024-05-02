CREATE TABLE IF NOT EXISTS clients(
      id UUID PRIMARY KEY,
      user_id UUID REFERENCES users(id) NOT NULL,
      grayscale_path TEXT,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      archived_at TIMESTAMP WITH TIME ZONE
);