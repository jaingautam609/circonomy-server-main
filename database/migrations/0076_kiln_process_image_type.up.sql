CREATE TYPE file_type AS ENUM ('image', 'video');

alter table klin_process_images
    add column file_type file_type default 'image'::file_type;