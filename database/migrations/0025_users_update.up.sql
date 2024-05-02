alter table users
    add column organisation_id uuid references organization(id);


alter table users_organisations
    add column real_organisation_id uuid references organization(id);