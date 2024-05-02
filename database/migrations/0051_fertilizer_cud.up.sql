Alter table fertilizers
    add column
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    add column archived_at TIMESTAMP WITH TIME ZONE;