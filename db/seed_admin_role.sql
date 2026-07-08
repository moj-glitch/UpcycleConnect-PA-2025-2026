insert into role (libelle, validation) values
    ('admin:general', '2')
on conflict (libelle) do nothing;
