ALTER TABLE crop_images
    DROP COLUMN crop_id;

ALTER TABLE crop_images
    ADD COLUMN farmer_crop_id uuid references farm_crops;