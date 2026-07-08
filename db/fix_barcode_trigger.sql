create or replace function generate_barcode_on_buy() returns trigger as $$
declare
    new_barcode BIGINT;
begin
    if NEW.etat = 'V' and NEW.acheteur is not null THEN
        new_barcode := NEW.vendeur # NEW.acheteur # NEW.annonce_id | (NEW.categorie::BIGINT << 48);
        NEW.barcode := new_barcode;
    elseif NEW.etat = 'V' and NEW.acheteur is null THEN
        NEW.barcode := 0;
    elseif NEW.etat != 'V' and NEW.acheteur is not null THEN
        NEW.etat = 'V';
        new_barcode := NEW.vendeur # NEW.acheteur # NEW.annonce_id | (NEW.categorie::BIGINT << 48);
        NEW.barcode := new_barcode;
    else
        NEW.barcode := null;
    end if;
    return NEW;
end $$ language plpgsql;

drop trigger if exists update_barcode_on_buy on annonce;

create trigger update_barcode_on_buy
before update on annonce for each row execute function generate_barcode_on_buy();
