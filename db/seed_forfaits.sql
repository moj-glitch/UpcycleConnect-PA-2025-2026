insert into forfait (libelle, mensuel, annuel, lifetime, prix) values
    ('Freemium', '0', '0', '1', 0),
    ('Pro Gratuit', '0', '0', '1', 0),
    ('Pro Payant', '1', '1', '0', 49.99)
on conflict (libelle) do nothing;
