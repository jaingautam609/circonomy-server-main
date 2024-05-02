CREATE TYPE  transportation_vehicle_type AS ENUM ('diesel', 'petrol', 'non_motorised');

ALTER TABLE farm_crops ADD COLUMN IF NOT EXISTS biomass_transportation_vehicle_type transportation_vehicle_type;