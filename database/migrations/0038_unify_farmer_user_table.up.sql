CREATE TABLE farmer_details
(
    user_id             uuid primary key references users,
    registration_status registration_status default 'verify_account'::registration_status
);

ALTER TABLE users
    ADD COLUMN age                 integer,
    ADD COLUMN gender              gender,
    ADD COLUMN aadhaar_no          text,
    ADD COLUMN aadhaar_no_image_id uuid references uploads;

ALTER TYPE account_type ADD VALUE 'farmer';

ALTER TABLE farmer_preferred_crop
    DROP farmer_id;
ALTER TABLE farmer_preferred_crop
    ADD COLUMN farmer_id uuid references users;

ALTER TABLE farms
    DROP farmer_id;
ALTER TABLE farms
    ADD COLUMN farmer_id uuid references users;