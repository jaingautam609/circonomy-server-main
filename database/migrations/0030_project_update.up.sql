alter table projects
    add column created_by uuid references users (id);

CREATE TYPE project_available_operation AS ENUM ('initial', 'addition', 'deduction');

CREATE TABLE IF NOT EXISTS project_available_changes
(
    id                  UUID PRIMARY KEY,
    created_by          UUID REFERENCES users (id)    NOT NULL,
    project_id          uuid references projects (id) not null,
    operation           project_available_operation   NOT NULL,
    amount              int,
    after_update_amount int,
    message             text,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at         TIMESTAMP WITH TIME ZONE
);