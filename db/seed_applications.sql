insert into etat_application (libelle) values
    ('Active'),
    ('Maintenance'),
    ('Retiree')
on conflict (libelle) do nothing;

insert into application (nom, description, prix, etat) values
    ('Tableau de bord avance', 'Tableau de bord avance pour le suivi de votre activite', 9.99, (select etat_application_id from etat_application where libelle = 'Active')),
    ('Analyse d''impact ecologique', 'Analyse d''impact ecologique detaillee de vos projets', 14.99, (select etat_application_id from etat_application where libelle = 'Active')),
    ('Statistiques materiaux', 'Statistiques detaillees sur les materiaux utilises', 7.99, (select etat_application_id from etat_application where libelle = 'Active')),
    ('Alertes priorisees', 'Alertes priorisees sur votre activite', 4.99, (select etat_application_id from etat_application where libelle = 'Active')),
    ('Fonctions IA', 'Fonctions d''intelligence artificielle', 19.99, (select etat_application_id from etat_application where libelle = 'Active'))
on conflict (nom) do nothing;

insert into role (libelle, validation) values
    ('admin:applications', '2')
on conflict (libelle) do nothing;
