Drop table farmer_rejected;

CREATE TABLE if not exists farmer_rejected
(
    id                      uuid primary key,
    user_id                 uuid references users,
    biomass_aggregator_id   uuid references biomass_aggregator,
    created_at              timestamp with time zone default now(),
    archived_at             timestamp with time zone
);