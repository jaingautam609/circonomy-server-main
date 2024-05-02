create table family
(
    id          uuid                     default gen_random_uuid() not null
        primary key,
    name        text                                               not null,
    created_by  uuid references users (id),
    created_at  timestamp with time zone default now(),
    archived_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS family_users
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id     uuid references users (id),
    family_id   uuid references family (id),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS family_invitations
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    email       varchar,
    family_id   uuid references users (id),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);
