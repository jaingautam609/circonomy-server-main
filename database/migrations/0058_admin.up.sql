create extension postgis;

CREATE TABLE IF NOT EXISTS biomass_aggregator
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    location    point,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS biomass_aggregator_manager
(
    id                    UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    biomass_aggregator_id uuid references biomass_aggregator (id) not null,
    manager_id            uuid references users (id)              not null,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at           TIMESTAMP WITH TIME ZONE
);


CREATE TABLE IF NOT EXISTS network
(
    id                    UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name                  TEXT                                    NOT NULL,
    biomass_aggregator_id uuid references biomass_aggregator (id) not null,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at           TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS network_manager
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    network_id  uuid references network (id) not null,
    manager_id  uuid references users (id)   not null,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);


CREATE TABLE IF NOT EXISTS klin
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        TEXT                         NOT NULL,
    network_id  uuid references network (id) not null,
    address     text,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS network_operator
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    klin_id     uuid references klin (id)  not null,
    operator_id uuid references users (id) not null,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

DROP TABLE farmers;

ALTER TABLE farmer_details
    ADD COLUMN network_id uuid references network (id);

alter table farm_crops add column klin_id uuid references klin(id);