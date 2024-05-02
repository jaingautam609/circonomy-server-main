ALTER type fertilizers RENAME TO ferti;

CREATE TABLE if not exists fertilizers
(
    id   uuid default gen_random_uuid() not null primary key,
    name text
);

ALTER TABLE crop_fertilizers
    DROP COLUMN fertilizer;

ALTER TABLE crop_fertilizers ADD COLUMN fertilizer_id uuid REFERENCES fertilizers;

DROP type ferti;