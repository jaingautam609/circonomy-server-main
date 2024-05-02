CREATE TYPE registration_status AS ENUM ('verify_account','create_account','add_farm','add_crop');

ALTER TABLE farmers ADD COLUMN if not exists registration_status registration_status default 'verify_account'::registration_status;