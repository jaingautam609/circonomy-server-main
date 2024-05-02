ALTER TABLE farm_crops
    ADD COLUMN IF NOT EXISTS cropping_started_at       timestamp with time zone DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS harvesting_started_at     timestamp with time zone,
    ADD COLUMN IF NOT EXISTS sun_drying_started_at     timestamp with time zone,
    ADD COLUMN IF NOT EXISTS transportation_started_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS production_started_at     timestamp with time zone,
    ADD COLUMN IF NOT EXISTS yield_quantity            int,
    ADD COLUMN IF NOT EXISTS yield_quantity_unit       area_unit,
    ADD COLUMN IF NOT EXISTS biomass_quantity          int,
    ADD COLUMN IF NOT EXISTS biomass_quantity_unit     area_unit;

