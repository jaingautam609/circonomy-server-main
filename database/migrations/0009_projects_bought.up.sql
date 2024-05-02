ALTER TABLE IF EXISTS projects RENAME COLUMN total_cost TO available;

CREATE TABLE IF NOT EXISTS projects_bought(
      id UUID PRIMARY KEY,
      user_id UUID REFERENCES users(id) NOT NULL,
      p_id UUID REFERENCES projects(id) NOT NULL,
      bought_by UUID REFERENCES users(id) NOT NULL,
      credits TEXT NOT NULL,
      bought_at_cost TEXT NOT NULL,
      bought_at_rate TEXT NOT NULL,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      archived_at TIMESTAMP WITH TIME ZONE
);
