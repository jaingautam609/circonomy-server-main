alter table projects alter column available type integer USING (available::integer);;
alter table projects alter column rate type integer USING (rate::integer);;