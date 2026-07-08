package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func planningHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodPut:
			rejoindreEvenement(w, r, token)
		case http.MethodGet:
			getPlanning(w, r, token)
		case http.MethodDelete:
			quitterEvenement(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

// rejoindreEvenement adds an existing catalog evenement to the caller's
// personal planning ("ajouter" in the CDC). It does not create a new
// evenement - only staff can do that (deposerEvenement).
func rejoindreEvenement(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.FormValue("evenement_id") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required field: evenement_id")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(
		context.Background(),
		"insert into participe (evenement_id, client_id) values ($1, $2) on conflict do nothing",
		r.FormValue("evenement_id"), token.ClientID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to join the evenement")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// quitterEvenement removes an evenement from the caller's personal planning
// ("supprimer" in the CDC - it removes the participation, not the event).
func quitterEvenement(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.FormValue("evenement_id") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required field: evenement_id")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(
		context.Background(),
		"delete from participe where evenement_id = $1 and client_id = $2",
		r.FormValue("evenement_id"), token.ClientID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to leave the evenement")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getPlanning lists the evenements the caller has joined - "ses prochains
// evenements" on the homepage when upcoming=1, or the full paginated list.
func getPlanning(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	fromRaw := r.URL.Query().Get("from")
	sizeRaw := r.URL.Query().Get("size")
	if fromRaw == "" || sizeRaw == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Pagination is required: provide from and size query params.")
		return
	}

	from, err := strconv.Atoi(fromRaw)
	if err != nil || from < 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from must be a non-negative integer.")
		return
	}

	size, err := strconv.Atoi(sizeRaw)
	if err != nil || size <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "size must be a strictly positive integer.")
		return
	}

	query := `select evenement.evenement_id, evenement.date, evenement.nom, evenement.description, evenement.statut
		from participe
		inner join evenement on evenement.evenement_id = participe.evenement_id
		where participe.client_id = $1`
	if r.URL.Query().Get("upcoming") == "1" {
		query += " and evenement.date >= now()"
	}
	query += " order by evenement.date asc limit $2 offset $3"

	rows, err := conn.Query(context.Background(), query, token.ClientID, size, from)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}
	defer rows.Close()

	entries := make([]map[string]any, 0)
	for rows.Next() {
		var evenementID int
		var date, nom, description, statut string

		if err := rows.Scan(&evenementID, &date, &nom, &description, &statut); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the database")
			return
		}

		entries = append(entries, map[string]any{
			"evenement_id": evenementID,
			"date":         date,
			"nom":          nom,
			"description":  description,
			"statut":       statut,
		})
	}

	jsonData, err := json.Marshal(entries)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal the planning")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
