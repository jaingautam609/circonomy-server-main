CREATE TABLE IF NOT EXISTS project_details_upload
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        text,
    upload_id   uuid unique references uploads (id),
    project_id  uuid references projects (id),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

alter table projects
    add column methodology text;
