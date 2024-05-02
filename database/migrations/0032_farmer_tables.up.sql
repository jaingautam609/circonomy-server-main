CREATE TYPE gender AS ENUM ('male', 'female', 'other');

CREATE TABLE IF NOT EXISTS farmers
(
    id                  UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name                TEXT NOT NULL,
    age                 INT,
    gender              gender,
    address             TEXT,
    phone_no            TEXT NOT NULL,
    profile_image_id    UUID REFERENCES uploads (id),
    aadhaar_no          TEXT,
    aadhaar_no_image_id UUID REFERENCES uploads (id),
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at         TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS farmers_phone_no_unique ON farmers (phone_no) WHERE archived_at IS NULL;

CREATE TYPE seasons AS ENUM ('zaid', 'kharif', 'rabi', 'all_seasons');

CREATE TABLE IF NOT EXISTS crops
(
    id            UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    crop_name     TEXT          NOT NULL,
    season        seasons       NOT NULL   DEFAULT 'all_seasons'::seasons,
    crop_image_id uuid
        references uploads (id) NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at   TIMESTAMP WITH TIME ZONE
);

CREATE TYPE area_unit AS ENUM ('bigha', 'hectare');

CREATE TYPE weight_unit AS ENUM ('kg', 'gm', 'sack');

CREATE TYPE farmer_app_video_content_type AS ENUM ('farming', 'biochar');

CREATE TYPE cropping_pattern AS ENUM ('mono_cropping', 'inter_cropping', 'crop_rotation', 'mixed_cropping' );

CREATE TYPE crop_stages AS ENUM ('cropping', 'harvesting', 'sun_drying', 'transportation', 'production', 'distribution');

CREATE TYPE fertilizers AS ENUM ('fertilizer_1', 'fertilizer_2', 'fertilizer_3', 'fertilizer_4', 'fertilizer_5');

CREATE TABLE IF NOT EXISTS farms
(
    id               UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    size             int,
    size_unit        area_unit,
    cropping_pattern cropping_pattern,
    farm_location    point,
    farm_images_ids  UUID[] not null,
    farmer_id        UUID REFERENCES farmers (id),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at      TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS farmer_preferred_crop
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    farmer_id   uuid REFERENCES farmers (id),
    crop_id     uuid REFERENCES crops (id),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS farm_crops
(
    id                       UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    crop_id                  uuid REFERENCES crops (id),
    farm_id                  uuid REFERENCES farms (id),
    crop_area                int,
    crop_area_unit           area_unit,
    cropping_pattern         cropping_pattern,
    crop_stage               crop_stages,
    seed_quantity            int,
    seed_quantity_unit       weight_unit,
    fertilizer               fertilizers,
    fertilizer_quantity      int,
    fertilizer_quantity_unit weight_unit,
    created_at               TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at               TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at              TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS crop_images
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    crop_id     uuid REFERENCES crops (id),
    crop_status crop_stages,
    image_id    uuid REFERENCES uploads (id),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS farm_app_content
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    title        text,
    description  text,
    content_type farmer_app_video_content_type,
    url          text,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at  TIMESTAMP WITH TIME ZONE
);

ALTER TABLE user_otp ADD COLUMN if not exists verified_at timestamp;

ALTER TYPE upload_type ADD VALUE if not exists 'farms';
ALTER TYPE upload_type ADD VALUE if not exists 'crops';
ALTER TYPE upload_type ADD VALUE if not exists 'farmer_educational_farming_video';
ALTER TYPE upload_type ADD VALUE if not exists 'farmer_educational_biochar_video';
