alter table users_organisations
    rename column org_id to old_org_id;


alter table users_organisations
    rename column real_organisation_id to organization_id;