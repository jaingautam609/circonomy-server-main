DROP TABLE IF EXISTS project_available_changes;

CREATE TABLE IF NOT EXISTS project_credits_operation
(
    id                     UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    created_by             UUID REFERENCES users (id)    NOT NULL,
    project_id             uuid references projects (id) not null,
    operation              project_available_operation   NOT NULL,
    amount                 int,
    after_operation_amount int,
    message                text,
    created_at             TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at            TIMESTAMP WITH TIME ZONE
);