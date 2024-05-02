alter table family_invitations drop column family_id;
alter table family_invitations add column family_id uuid references family(id);