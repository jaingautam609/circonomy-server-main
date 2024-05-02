CREATE TABLE IF NOT EXISTS klin_biomass
(
    id               UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    klin_id          uuid references klin (id)  not null,
    crop_id          uuid references crops (id) not null,
    current_quantity numeric                  default 0,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at      TIMESTAMP WITH TIME ZONE
);

alter table klin
    add column biochar_quantity numeric default 0;

CREATE TABLE IF NOT EXISTS klin_process
(
    id               UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    klin_id          uuid references klin (id)  not null,
    end_time         timestamptz,
    crop_id          uuid references crops (id) not null,
    biomass_quantity numeric                  default 0,
    biochar_quantity numeric                  default 0,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at      TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS klin_process_images
(
    id              UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    klin_process_id uuid references klin_process (id) not null,
    upload_id       uuid references uploads (id),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at     TIMESTAMP WITH TIME ZONE
);