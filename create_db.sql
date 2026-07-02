
CREATE TABLE client (
    client_id SERIAL PRIMARY KEY,
    nom varchar(64) NOT NULL,
    prenom varchar(32) NOT NULL,
    email varchar(128) NOT NULL,
    password varchar(255) NOT NULL,
    telephone varchar(12) NOT NULL check (telephone ~ '^\+[0-9]{11}$'),
    adresse varchar(128) NOT NULL,
    adresse2 varchar(128),
    complement_adresse varchar(128),
    code_postal varchar(5) NOT NULL check (code_postal ~ '^[0-9]{5}$'),
    ville varchar(64) NOT NULL,
    score BIGINT NOT NULL DEFAULT 0,
    siret CHAR(14) NOT NULL UNIQUE check (siret ~ '^[0-9]{14}$')
);

create table categorie_annonce(
    categorie_id SERIAL PRIMARY KEY,
    libelle varchar(32) NOT NULL UNIQUE
);

create table norme(
    norme_id SERIAL PRIMARY KEY,
    code varchar(16) NOT NULL UNIQUE
);

create table materieau(
    materieau_id SERIAL PRIMARY KEY,
    nom varchar(32) NOT NULL UNIQUE,
    densite real NOT NULL,
    prix_kg real NOT NULL check (prix_kg > 0),
    durete real NOT NULL,
    fonte_degree real NOT NULL check (fonte_degree > -273.15),
    elasticite_young real NOT NULL,
    resistance_torsion real NOT NULL,
    resistance_compression real NOT NULL,
    resilience real NOT NULL,
    conductivite_thermique real NOT NULL,
    conductivite_electrique real NOT NULL,
    dilatation_thermique real NOT NULL,
    porosite real NOT NULL,
    opacite real NOT NULL,
    magnetisme real NOT NULL,
    accoustique real NOT NULL,
    recyclable char(1) NOT NULL check (recyclable in ('1', '0')),
    empreinte_co2 real NOT NULL,
    toxicite real NOT NULL
);

create table norme_materieau(
    norme_id INTEGER NOT null REFERENCES norme(norme_id),
    materieau_id INTEGER NOT NULL REFERENCES materieau(materieau_id),
    PRIMARY KEY (norme_id, materieau_id)
);
    
create table annonce(
    annonce_id SERIAL PRIMARY KEY,
    vendeur INTEGER NOT NULL REFERENCES client(client_id),
    acheteur INTEGER REFERENCES client(client_id),
    categorie INTEGER NOT NULL REFERENCES categorie_annonce(categorie_id),
    titre varchar(64) NOT NULL,
    prix FLOAT NOT NULL check (prix >= 0),
    description varchar(512) NOT NULL,
    etat CHAR(1) NOT NULL check (etat in ('D', 'V', 'S', 'F', 'J', 'E', 'K')), -- D: Disponible, V: Vendue, S: Supprimée, F: Finalisée, J: Déposée par le vendeur, E: En cours d'acheminement, K: acheminé
    taxe NUMERIC(1, 2) NOT NULL DEFAULT 0.20 check (taxe >= 0 and taxe <= 1),
    image VARCHAR(128) NOT NULL,
    barcode BIGINT,
    date_publication TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table utilise(
    annonce_id INTEGER NOT NULL REFERENCES annonce(annonce_id),
    materieau_id INTEGER NOT NULL REFERENCES materieau(materieau_id),
    PRIMARY KEY (annonce_id, materieau_id)
);

create or replace function generate_barcode_on_buy() returns trigger as $$
declare
    new_barcode BIGINT;
begin
    if NEW.etat = 'V' and NEW.acheteur is not null THEN
        new_barcode := NEW.vendeur ^ NEW.acheteur ^ NEW.annonce_id | (NEW.categorie << 48);
        NEW.barcode := new_barcode;
    elseif NEW.etat = 'V' and NEW.acheteur is null THEN
        new_barcode := 0;
        NEW.barcode := new_barcode;
    elseif NEW.etat != 'V' and NEW.acheteur is not null THEN
        NEW.etat = 'V';
        new_barcode := NEW.vendeur ^ NEW.acheteur ^ NEW.annonce_id | (NEW.categorie << 48);
        NEW.barcode := new_barcode;
    else
        NEW.barcode := null;
    end if;
end $$ language plpgsql;

create or replace trigger update_barcode_on_buy
AFTER UPDATE ON annonce for each row execute function generate_barcode_on_buy();

create or replace VIEW annonces_per_category AS
SELECT libelle AS categorie, annonce_id, barcode, titre, acheteur, prix, etat, description 
FROM annonce INNER JOIN categorie_annonce 
on annonce.categorie = categorie_annonce.categorie_id 
GROUP BY categorie_annonce.libelle, annonce_id 
ORDER BY annonce.date_publication DESC;

create table message_annonce(
    message_annonce_id SERIAL PRIMARY KEY,
    annonce_id INTEGER NOT NULL REFERENCES annonce(annonce_id),
    client_id INTEGER NOT NULL REFERENCES client(client_id),
    message varchar(128) NOT NULL,
    etat CHAR(1) NOT NULL check (etat in ('L', 'N')) -- L: Lu, N: Non lu
);

create or replace VIEW messages_per_annonce AS
SELECT annonce.annonce_id, message_annonce_id, client_id, message, message_annonce.etat
FROM message_annonce INNER JOIN annonce
ON message_annonce.annonce_id = annonce.annonce_id;

create table categorie_thread(
    categorie_thread_id SERIAL PRIMARY KEY,
    libelle varchar(32) NOT NULL UNIQUE
);

create table thread(
    thread_id SERIAL PRIMARY KEY,
    categorie_thread INTEGER NOT NULL REFERENCES categorie_thread(categorie_thread_id),
    titre varchar(64) NOT NULL,
    message varchar(512) NOT NULL,
    resolu char(1) NOT NULL check (resolu in ('1', '0')), -- 1: Oui, 0: Non
    date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table message_thread(
    message_thread_id SERIAL PRIMARY KEY,
    thread_id INTEGER NOT NULL REFERENCES thread(thread_id),
    client_id INTEGER NOT NULL REFERENCES client(client_id),
    message varchar(128) NOT NULL,
    parent INTEGER REFERENCES message_thread(message_thread_id),
    date_envoi TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table categorie_staff(
    categorie_staff_id SERIAL PRIMARY KEY,
    libelle varchar(64) NOT NULL UNIQUE
);

create table evenement(
    evenement_id SERIAL PRIMARY KEY,
    date TIMESTAMP NOT NULL,
    nom VARCHAR(64) NOT NULL,
    description VARCHAR(512) NOT NULL,
    statut CHAR(1) NOT NULL check (statut in ('P', 'C', 'A')), -- P: Planifié, C: En cours, A: Annulé
    categorie INTEGER NOT NULL REFERENCES categorie_staff(categorie_staff_id)
);

create table participe(
    evenement_id INTEGER NOT NULL REFERENCES evenement(evenement_id),
    client_id INTEGER NOT NULL REFERENCES client(client_id),
    PRIMARY KEY (evenement_id, client_id)
)

create table tutoriel(
    tutoriel_id SERIAL PRIMARY KEY,
    titre VARCHAR(64) NOT NULL,
    date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    article TEXT NOT NULL,
    video VARCHAR(128) NOT NULL,
    categorie INTEGER NOT NULL REFERENCES categorie_staff(categorie_staff_id)
);

create table forfait(
    forfait_id SERIAL PRIMARY KEY,
    libelle VARCHAR(32) NOT NULL UNIQUE,
    mensuel CHAR(1) NOT NULL check (mensuel in ('1', '0')), -- 1: Oui, 0: Non
    annuel CHAR(1) NOT NULL check (annuel in ('1', '0')), -- 1: Oui, 0: Non
    lifetime CHAR(1) NOT NULL check (lifetime in ('1', '0')), -- 1: Oui, 0: Non
    prix FLOAT NOT NULL check (prix >= 0)
);

create table souscrit(
    client_id INTEGER NOT NULL REFERENCES client(client_id),
    forfait_id INTEGER NOT NULL REFERENCES forfait(forfait_id),
    date_debut TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_fin TIMESTAMP,
    PRIMARY KEY (client_id, forfait_id)
);

create table etat_application(
    etat_application_id SERIAL PRIMARY KEY,
    libelle VARCHAR(32) NOT NULL UNIQUE
);

create table application(
    application_id SERIAL PRIMARY KEY,
    nom VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(512) NOT NULL,
    prix FLOAT NOT NULL check (prix >= 0),
    etat INTEGER NOT NULL REFERENCES etat_application(etat_application_id)
);

create table role(
    role_id SERIAL PRIMARY KEY,
    libelle VARCHAR(32) NOT NULL UNIQUE,
    validation char(1) NOT NULL check (validation in ('2', '1', '0')) -- 2: Validation automatique, 1: Validé, 0: Non validé
);

create table grant_app(
    application_id INTEGER NOT NULL REFERENCES application(application_id),
    role_id INTEGER NOT NULL REFERENCES role(role_id),
);

create table possede(
    client_id INTEGER NOT NULL REFERENCES client(client_id),
    role_id INTEGER NOT NULL REFERENCES role(role_id),
    PRIMARY KEY (client_id, role_id)
);

create table accede(
    client_id INTEGER NOT NULL REFERENCES client(client_id),
    application_id INTEGER NOT NULL REFERENCES application(application_id),
    achetee CHAR(1) NOT NULL check (achetee in ('1', '0')), -- 1: Oui, 0: Non
    PRIMARY KEY (client_id, application_id)
);

create function empeche_achat_si_pas_en_freemium() returns trigger as $$
declare
    client_forfait RECORD;
    has_freemium BOOLEAN := FALSE;
begin
    for client_forfait IN SELECT * FROM souscrit WHERE client_id = NEW.client_id LOOP
        if client_forfait.forfait_id = (SELECT forfait_id FROM forfait WHERE libelle = 'Freemium') THEN
            has_freemium := TRUE;
            exit;
        end if;
    END LOOP;

    if NOT has_freemium THEN
        RAISE EXCEPTION 'Achat impossible : le client n''a pas de forfait Freemium actif.';
    end if;

    return NEW;
end $$ language plpgsql;

create trigger check_freemium_before_purchase
BEFORE INSERT ON accede for each row execute function empeche_achat_si_pas_en_freemium();

create table token(
    token_id SERIAL PRIMARY KEY,
    client_id INTEGER NOT NULL REFERENCES client(client_id),
    token_value VARCHAR(256) NOT NULL UNIQUE,
    date_expiration TIMESTAMP NOT NULL,
    active char(1) NOT NULL check (active in ('1', '0')) DEFAULT '1', -- 1: Actif, 0: Inactif
    revoked char(1) NOT NULL check (revoked in ('1', '0')) DEFAULT '0' -- 1: Révoqué, 0: Non révoqué
);

create table entreprise (
    entreprise_id SERIAL PRIMARY KEY,
    nom VARCHAR(64) NOT NULL UNIQUE,
    adresse VARCHAR(128) NOT NULL,
    code_postal VARCHAR(5) NOT NULL check (code_postal ~ '^[0-9]{5}$'),
    ville VARCHAR(64) NOT NULL,
    siret CHAR(14) NOT NULL UNIQUE check (siret ~ '^[0-9]{14}$')
);

create table employe(
    entreprise_id INTEGER NOT NULL REFERENCES entreprise(entreprise_id),
    manager_id INTEGER REFERENCES client(client_id),
    client_id INTEGER NOT NULL REFERENCES client(client_id),
    PRIMARY KEY (entreprise_id, client_id)
);

create table categorie_contrat(
    categorie_contrat_id SERIAL PRIMARY KEY,
    libelle VARCHAR(32) NOT NULL UNIQUE
);

create table contrat(
    contrat_id SERIAL PRIMARY KEY,
    entreprise_id INTEGER NOT null REFERENCES entreprise(entreprise_id),
    categorie_contrat_id INTEGER REFERENCES categorie_contrat(categorie_contrat_id),
    description VARCHAR(512) NOT NULL,
    depense FLOAT NOT NULL check (depense >= 0),
    gain FLOAT NOT NULL check (gain >= 0),
    pdf VARCHAR(128) NOT NULL,
    date_debut TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_fin TIMESTAMP
);

create table tiers(
    tiers_id SERIAL PRIMARY KEY,
    nom VARCHAR(64) NOT NULL UNIQUE,
    adresse VARCHAR(128) NOT NULL,
    code_postal VARCHAR(5) NOT NULL check (code_postal ~ '^[0-9]{5}$'),
    ville VARCHAR(64) NOT NULL,
    siret CHAR(14) UNIQUE check (siret ~ '^[0-9]{14}$')
);

create table categorie_projet(
    categorie_projet_id SERIAL PRIMARY KEY,
    libelle VARCHAR(32) NOT NULL UNIQUE
);

create table projet(
    projet_id SERIAL PRIMARY KEY,
    categorie_projet INTEGER NOT NULL REFERENCES categorie_projet(categorie_projet_id),
    entreprise_id INTEGER NOT NULL REFERENCES entreprise(entreprise_id),
    titre VARCHAR(64) NOT NULL UNIQUE,
    image VARCHAR(128) NOT NULL,
    video VARCHAR(128) NOT NULL,
    texte text NOT NULL,
    date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

