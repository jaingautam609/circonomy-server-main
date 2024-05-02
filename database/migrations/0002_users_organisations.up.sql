CREATE TABLE IF NOT EXISTS users (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name TEXT NOT NULL,
     email TEXT UNIQUE,
     password TEXT,
     number TEXT UNIQUE,
     location TEXT,
     account_type account_type,
     size business_size,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
     archived_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS users_id_index ON users(id);
CREATE INDEX IF NOT EXISTS users_email_index ON users(email);
CREATE INDEX IF NOT EXISTS users_name_index ON users(name);

CREATE TABLE IF NOT EXISTS users_organisations (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   user_id uuid references users(id),
   org_id uuid references users(id),
   created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
   archived_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS org_id_index ON users_organisations(org_id);
