CREATE TYPE otp_type AS ENUM ('email','number');

CREATE TABLE IF NOT EXISTS user_otp(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    input TEXT,
    otp TEXT,
    type otp_type,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);