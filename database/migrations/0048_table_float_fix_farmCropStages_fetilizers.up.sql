ALTER TABLE if exists farms
    ADD COLUMN landmark text,
    alter COLUMN size type numeric(8,2);

CREATE TABLE if not exists crop_fertilizers
(
    id                       uuid default gen_random_uuid() not null primary key,
    fertilizer               fertilizers,
    fertilizer_quantity      numeric(8,2),
    fertilizer_quantity_unit weight_unit,
    farm_crop_id             uuid references farm_crops
);

CREATE TABLE if not exists farm_crop_stages
(
    id            uuid default gen_random_uuid() not null primary key,
    stage         crop_stages,
    starting_time timestamp with time zone,
    farm_crop_id  uuid references farm_crops
);

ALTER TABLE if exists farm_crops
    drop COLUMN cropping_started_at,
    drop COLUMN harvesting_started_at,
    drop COLUMN sun_drying_started_at,
    drop COLUMN transportation_started_at,
    drop COLUMN production_started_at,
    drop COLUMN fertilizer,
    drop COLUMN fertilizer_quantity,
    drop COLUMN fertilizer_quantity_unit,
    alter COLUMN crop_area type numeric(8,2),
    alter COLUMN seed_quantity type numeric(8,2),
    alter COLUMN yield_quantity type numeric(8,2),
    alter COLUMN biomass_quantity type numeric(8,2);
