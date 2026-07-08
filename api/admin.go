package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func adminClientsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		if token.Scope == nil || !contains(token.Scope, "admin:general") {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only general administrators can access this")
			return
		}
		switch r.Method {
		case http.MethodGet:
			getClientsAdmin(w, r)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func adminRolesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		if token.Scope == nil || !contains(token.Scope, "admin:general") {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only general administrators can access this")
			return
		}
		switch r.Method {
		case http.MethodGet:
			getRolesAdmin(w, r)
		case http.MethodPut:
			grantRoleAdmin(w, r)
		case http.MethodDelete:
			revokeRoleAdmin(w, r)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func getClientsAdmin(w http.ResponseWriter, r *http.Request) {
	fromRaw := r.URL.Query().Get("from")
	sizeRaw := r.URL.Query().Get("size")
	if fromRaw == "" || sizeRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from and size are required")
		return
	}

	from, err := strconv.Atoi(fromRaw)
	if err != nil || from < 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from must be >= 0")
		return
	}

	size, err := strconv.Atoi(sizeRaw)
	if err != nil || size <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "size must be > 0")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_ADMIN_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	query := `select client.client_id, client.nom, client.prenom, client.email, client.score,
		coalesce(string_agg(role.libelle, ' ' order by role.libelle), '')
		from client
		left join possede on possede.client_id = client.client_id
		left join role on role.role_id = possede.role_id`
	args := []any{}
	argPos := 1
	if q := r.URL.Query().Get("q"); q != "" {
		query += " where client.email ilike $" + strconv.Itoa(argPos) + " or client.nom ilike $" + strconv.Itoa(argPos)
		args = append(args, "%"+q+"%")
		argPos++
	}
	query += " group by client.client_id order by client.client_id offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(context.Background(), query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query clients")
		return
	}
	defer rows.Close()

	clients := make([]map[string]any, 0)
	for rows.Next() {
		var clientID int
		var score int64
		var nom, prenom, email, roles string

		if err := rows.Scan(&clientID, &nom, &prenom, &email, &score, &roles); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan clients")
			return
		}

		clients = append(clients, map[string]any{
			"client_id": clientID,
			"nom":       nom,
			"prenom":    prenom,
			"email":     email,
			"score":     score,
			"roles":     roles,
		})
	}

	writeJSON(w, http.StatusOK, clients)
}

func getRolesAdmin(w http.ResponseWriter, r *http.Request) {
	conn, err := pgx.Connect(context.Background(), DATABASE_ADMIN_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select role_id, libelle from role order by libelle")
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query roles")
		return
	}
	defer rows.Close()

	roles := make([]map[string]any, 0)
	for rows.Next() {
		var roleID int
		var libelle string
		if err := rows.Scan(&roleID, &libelle); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan roles")
			return
		}
		roles = append(roles, map[string]any{"role_id": roleID, "libelle": libelle})
	}

	writeJSON(w, http.StatusOK, roles)
}

func grantRoleAdmin(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("client_id") == "" || r.FormValue("libelle") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "client_id and libelle are required")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_ADMIN_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(
		context.Background(),
		"insert into possede (client_id, role_id) select $1, role_id from role where libelle = $2 on conflict do nothing",
		r.FormValue("client_id"), r.FormValue("libelle"),
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to grant the role")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func revokeRoleAdmin(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("client_id") == "" || r.FormValue("libelle") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "client_id and libelle are required")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_ADMIN_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(
		context.Background(),
		"delete from possede using role where possede.role_id = role.role_id and possede.client_id = $1 and role.libelle = $2",
		r.FormValue("client_id"), r.FormValue("libelle"),
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to revoke the role")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
