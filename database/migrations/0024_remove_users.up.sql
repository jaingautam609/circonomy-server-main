create table organization
(
    id          uuid                     default gen_random_uuid() not null
        primary key,
    name        text                                               not null,
    user_id     uuid references users (id),
    created_at  timestamp with time zone default now(),
    archived_at timestamp with time zone
);

