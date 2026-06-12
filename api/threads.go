package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func threadsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodPut:
			deposerThread(w, r, token)
		case http.MethodGet:
			getThread(w, r, token)
		case http.MethodPatch:
			modifierThread(w, r, token)
		case http.MethodDelete:
			supprimerThread(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func messagesThreadHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodPut:
			envoyerMessageThread(w, r, token)
		case http.MethodGet:
			consulterMessagesThread(w, r, token)
		case http.MethodPatch:
			modifierMessageThread(w, r, token)
		case http.MethodDelete:
			supprimerMessageThread(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func messagesThreadChildrenHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodGet:
			getMessagesThreadChildren(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func deposerThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.FormValue("categorie_thread") == "" || r.FormValue("titre") == "" || r.FormValue("message") == "" || r.FormValue("resolu") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "categorie_thread, titre, message and resolu are required")
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
		"insert into thread (categorie_thread, titre, message, resolu) values ($1, $2, $3, $4)",
		r.FormValue("categorie_thread"),
		r.FormValue("titre"),
		r.FormValue("message"),
		r.FormValue("resolu"),
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to insert the thread")
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = token.ClientID
}

func getThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.URL.Query().Get("from") != "" && r.URL.Query().Get("size") != "" {
		conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
			return
		}
		defer conn.Close(context.Background())

		rows, err := conn.Query(context.Background(), "select thread_id, categorie_thread, titre, message, resolu, date_creation from thread order by date_creation desc limit $1 offset $2", r.URL.Query().Get("size"), r.URL.Query().Get("from"))
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
			return
		}
		defer rows.Close()

		threads := make([]Thread, 0)
		for rows.Next() {
			var thread Thread
			err = rows.Scan(&thread.ThreadID, &thread.Categorie, &thread.Titre, &thread.Message, &thread.Resolu, &thread.DateCreate)
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the database")
				return
			}
			threads = append(threads, thread)
		}

		jsonData, err := json.Marshal(threads)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal threads")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else if r.URL.Query().Get("id") != "" {
		conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
			return
		}
		defer conn.Close(context.Background())

		var thread Thread
		err = conn.QueryRow(context.Background(), "select thread_id, categorie_thread, titre, message, resolu, date_creation from thread where thread_id = $1", r.URL.Query().Get("id")).Scan(&thread.ThreadID, &thread.Categorie, &thread.Titre, &thread.Message, &thread.Resolu, &thread.DateCreate)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
			return
		}

		jsonData, err := json.Marshal(thread)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal the thread")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "You must provide either the id of the thread or the from and size query params.")
	}

	_ = token.ClientID
}

func modifierThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	isAdmin := token.Scope != nil && contains(token.Scope, "public:administrateur_des_threads")
	if !isAdmin {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to modify this thread")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "update thread set categorie_thread = $1, titre = $2, message = $3, resolu = $4 where thread_id = $5", r.FormValue("categorie_thread"), r.FormValue("titre"), r.FormValue("message"), r.FormValue("resolu"), r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the thread")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func supprimerThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	isAdmin := token.Scope != nil && contains(token.Scope, "public:administrateur_des_threads")

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	var ownerID int
	err = conn.QueryRow(context.Background(), "select client_id from message_thread where thread_id = $1 order by message_thread_id asc limit 1", r.FormValue("id")).Scan(&ownerID)
	if err != nil {
		ownerID = 0
	}

	if !isAdmin && token.ClientID != ownerID {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to delete this thread")
		return
	}

	tx, err := conn.Begin(context.Background())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to start transaction")
		return
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "delete from message_thread where thread_id = $1", r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete thread messages")
		return
	}

	_, err = tx.Exec(context.Background(), "delete from thread where thread_id = $1", r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete the thread")
		return
	}

	err = tx.Commit(context.Background())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to commit transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func consulterMessagesThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.URL.Query().Get("thread_id") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "thread_id query param is required")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select message_thread_id, thread_id, client_id, message, parent, date_envoi from message_thread where thread_id = $1 and parent is null order by date_envoi asc", r.URL.Query().Get("thread_id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}
	defer rows.Close()

	messages := make([]MessageThread, 0)
	for rows.Next() {
		var message MessageThread
		err = rows.Scan(&message.MessageThreadID, &message.ThreadID, &message.ClientID, &message.Message, &message.Parent, &message.DateEnvoi)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the database")
			return
		}
		messages = append(messages, message)
	}

	jsonPayload, err := json.Marshal(messages)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal messages")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonPayload)

	_ = token.ClientID
}

func getMessagesThreadChildren(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	threadID := r.URL.Query().Get("thread_id")
	parentID := r.URL.Query().Get("parent_id")
	if threadID == "" || parentID == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "thread_id and parent_id query params are required")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select message_thread_id, thread_id, client_id, message, parent, date_envoi from message_thread where thread_id = $1 and parent = $2 order by date_envoi asc", threadID, parentID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}
	defer rows.Close()

	messages := make([]MessageThread, 0)
	for rows.Next() {
		var message MessageThread
		err = rows.Scan(&message.MessageThreadID, &message.ThreadID, &message.ClientID, &message.Message, &message.Parent, &message.DateEnvoi)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the database")
			return
		}
		messages = append(messages, message)
	}

	jsonPayload, err := json.Marshal(messages)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal messages")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonPayload)

	_ = token.ClientID
}

func envoyerMessageThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.FormValue("thread_id") == "" || r.FormValue("message") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "thread_id and message are required")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	if r.FormValue("parent") == "" {
		_, err = conn.Exec(context.Background(), "insert into message_thread (thread_id, client_id, message) values ($1, $2, $3)", r.FormValue("thread_id"), token.ClientID, r.FormValue("message"))
	} else {
		_, err = conn.Exec(context.Background(), "insert into message_thread (thread_id, client_id, message, parent) values ($1, $2, $3, $4)", r.FormValue("thread_id"), token.ClientID, r.FormValue("message"), r.FormValue("parent"))
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to insert the message")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func supprimerMessageThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	var ownerID int
	err = conn.QueryRow(context.Background(), "select client_id from message_thread where message_thread_id = $1", r.FormValue("id")).Scan(&ownerID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}

	isAdmin := token.Scope != nil && contains(token.Scope, "public:administrateur_des_threads")
	if token.ClientID != ownerID && !isAdmin {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to delete this message")
		return
	}

	_, err = conn.Exec(context.Background(), "delete from message_thread where message_thread_id = $1", r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete the message")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func modifierMessageThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	writeAPIError(w, http.StatusForbidden, "forbidden", "Messages cannot be modified once sent.")
	_ = token.ClientID
	_ = r.URL.String()
}
