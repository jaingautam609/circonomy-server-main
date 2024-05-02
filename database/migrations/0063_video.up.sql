CREATE TYPE video_tag AS ENUM ('farming','biochar');

CREATE TABLE IF NOT EXISTS video
(
    id                 UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    title              text,
    description        text,
    url                text,
    video_tag          video_tag                default 'farming'::video_tag,
    thumbnail_image_id uuid references uploads (id),
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at        TIMESTAMP WITH TIME ZONE
);