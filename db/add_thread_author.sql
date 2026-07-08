alter table thread add column client_id INTEGER REFERENCES client(client_id);
