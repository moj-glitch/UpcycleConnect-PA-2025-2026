create table actor(
    tiers_id SERIAL REFERENCES tiers(tiers_id),
    contrat_id SERIAL REFERENCES contrat(contrat_id),
    PRIMARY KEY (tiers_id, contrat_id)
);
