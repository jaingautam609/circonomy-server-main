alter table projects
    alter column capacity type integer using (capacity::integer);

alter table contacts
    add column linkedin_link text;