CREATE TABLE IF NOT EXISTS sessions (
    id            UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id       UUID REFERENCES users(id) NOT NULL,
    session_token TEXT                       NOT NULL,
    expiry_time   TIMESTAMP WITH TIME ZONE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);