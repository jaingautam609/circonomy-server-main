CREATE TABLE if not exists farmer_rejected
(
    user_id                 uuid not null primary key references users,
    biomass_aggregator_id   uuid references biomass_aggregator,
    created_at              timestamp with time zone default now(),
    archived_at             timestamp with time zone
);