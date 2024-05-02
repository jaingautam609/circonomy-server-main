alter table contacts
    add column designation text;
alter table contacts
    add column phone integer;

CREATE TYPE certificate_status AS ENUM ('valid','expired');

alter table certificates
    add column status certificate_status default 'valid'::certificate_status;