package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func applicationsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodGet:
			getApplications(w, r, token)
		case http.MethodPatch:
			modifierApplicationPrix(w, r, token)
		case http.MethodPut:
			acheterApplication(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func getApplications(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(
		context.Background(),
		`select application.application_id, application.nom, application.description, application.prix, etat_application.libelle,
		coalesce((select achetee from accede where accede.application_id = application.application_id and accede.client_id = $1), '0')
		from application
		inner join etat_application on etat_application.etat_application_id = application.etat
		order by application.application_id`,
		token.ClientID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query applications")
		return
	}
	defer rows.Close()

	applications := make([]map[string]any, 0)
	for rows.Next() {
		var applicationID int
		var nom, description, etat, achetee string
		var prix float64

		if err := rows.Scan(&applicationID, &nom, &description, &prix, &etat, &achetee); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan applications")
			return
		}

		applications = append(applications, map[string]any{
			"application_id": applicationID,
			"nom":            nom,
			"description":    description,
			"prix":           prix,
			"etat":           etat,
			"achetee":        achetee == "1",
		})
	}

	writeJSON(w, http.StatusOK, applications)
}

// modifierApplicationPrix is the back-office price management endpoint,
// restricted to the admin:applications scope.
func modifierApplicationPrix(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if token.Scope == nil || !contains(token.Scope, "admin:applications") {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only application administrators can update prices")
		return
	}

	if r.FormValue("id") == "" || r.FormValue("prix") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required fields: id, prix")
		return
	}

	prix, err := strconv.ParseFloat(r.FormValue("prix"), 64)
	if err != nil || prix < 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "prix must be a non-negative number")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	tag, err := conn.Exec(context.Background(), "update application set prix = $1 where application_id = $2", prix, r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the price")
		return
	}
	if tag.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Application not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// acheterApplication records a purchase in accede. It is called by the
// frontend only after Stripe confirms the checkout session is paid. Note:
// the check_freemium_before_purchase trigger requires the buyer to hold the
// 'Freemium' forfait.
func acheterApplication(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.FormValue("application_id") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required field: application_id")
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
		"insert into accede (client_id, application_id, achetee) values ($1, $2, '1') on conflict (client_id, application_id) do update set achetee = '1'",
		token.ClientID, r.FormValue("application_id"),
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to record the purchase")
		return
	}

	w.WriteHeader(http.StatusCreated)
}
