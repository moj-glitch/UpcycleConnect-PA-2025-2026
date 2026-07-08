-- create_db.sql doesn't seed the role table, but the app's grantRoles()
-- (oauth's createAccount for account_type=entreprise, and api's RH
-- create-employee-account endpoint) looks up role_id by libelle and
-- silently skips anything it can't find. Without these rows, a freshly
-- registered CEO ends up with zero usable scopes even though registration
-- itself succeeds.
--
-- validation='2' (auto-validated) since these are granted directly by
-- trusted server-side logic, not through any pending-approval workflow.
--
-- Note: libelle is VARCHAR(32) - three names had to be shortened to fit
-- (see api/annonces.go, api/contrats.go, api/threads.go, which reference
-- the shortened forms directly):
--   pro:support_contrat_administrateur  -> pro:contrat_administrateur
--   public:administrateur_des_annonces  -> public:admin_annonces
--   public:administrateur_des_threads   -> public:admin_threads

insert into role (libelle, validation) values
    ('pro:pdg', '2'),
    ('pro:entreprise_manager', '2'),
    ('pro:entreprise_administrateur', '2'),
    ('pro:rh', '2'),
    ('pro:rh_support', '2'),
    ('pro:manager', '2'),
    ('pro:gestionnaire_contrats', '2'),
    ('pro:contrat_administrateur', '2'),
    ('pro:project_manager', '2'),
    ('pro:project_administrateur', '2'),
    ('public:admin_annonces', '2'),
    ('public:admin_threads', '2'),
    ('evenements:manager', '2'),
    ('tutorials:content_manager', '2')
on conflict (libelle) do nothing;
