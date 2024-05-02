CREATE TABLE if not exists enquiry
(
    id         uuid        default gen_random_uuid() not null primary key,
    first_name text,
    last_name  text,
    email      text,
    created_at timestamptz default now()
);