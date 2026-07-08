package main

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func scoreHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
			return
		}
		getScore(w, r, token)
	}
}

func getScore(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	var score int64
	err = conn.QueryRow(context.Background(), "select score from client where client_id = $1", token.ClientID).Scan(&score)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the score")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"client_id": token.ClientID,
		"score":     score,
	})
}
