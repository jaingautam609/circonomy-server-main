alter table crop_fertilizers add column if not exists created_at timestamptz default now();
alter table crop_fertilizers add column if not exists archived_at timestamptz;