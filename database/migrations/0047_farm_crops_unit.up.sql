ALTER TABLE farm_crops
    DROP COLUMN if exists yield_quantity_unit;

ALTER TABLE farm_crops
    ADD COLUMN if not exists yield_quantity_unit weight_unit;

ALTER TABLE farm_crops
    DROP COLUMN if exists biomass_quantity_unit;

ALTER TABLE farm_crops
    ADD COLUMN if not exists biomass_quantity_unit weight_unit;