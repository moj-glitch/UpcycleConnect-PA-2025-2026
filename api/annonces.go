package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func annoncesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodPut:
			deposerAnnonce(w, r, token)
		case http.MethodGet:
			getAnnonce(w, r, token)
		case http.MethodDelete:
			supprimerAnnonce(w, r, token)
		case http.MethodPatch:
			modifierAnnonce(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func messagesAnnonceHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodPut:
			envoyerMessageAnnonce(w, r, token)
		case http.MethodGet:
			consulterMessagesAnnonce(w, r, token)
		case http.MethodPatch:
			modifierMessageAnnonce(w, r, token)
		case http.MethodDelete:
			supprimerMessageAnnonce(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func categoriesAnnoncesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodGet:
			getCategorieAnnonce(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func deposerAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	tx, err := conn.Begin(context.Background())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to start transaction")
		return
	}
	defer tx.Rollback(context.Background())

	var annonceID int
	err = tx.QueryRow(
		context.Background(),
		"insert into annonce (vendeur, categorie, titre, prix, description, etat, taxe, image) values ($1, $2, $3, $4, $5, $6, $7, $8) returning annonce_id",
		token.ClientID,
		r.FormValue("categorie"),
		r.FormValue("titre"),
		r.FormValue("prix"),
		r.FormValue("description"),
		r.FormValue("etat"),
		r.FormValue("taxe"),
		r.FormValue("image"),
	).Scan(&annonceID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to insert the annonce")
		return
	}

	for _, materieauID := range r.Form["materiaux"] {
		if materieauID == "" {
			continue
		}
		_, err = tx.Exec(context.Background(), "insert into utilise (annonce_id, materieau_id) values ($1, $2)", annonceID, materieauID)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to record the materiau used")
			return
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to commit transaction")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"annonce_id": annonceID})
}

func achatAnnonceHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		if r.Method != http.MethodPatch {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
			return
		}
		acheterAnnonce(w, r, token)
	}
}

// acheterAnnonce lets a buyer claim an available annonce. It only ever
// touches etat and acheteur - the barcode column is computed by the
// generate_barcode_on_buy() trigger in create_db.sql, never by the API.
func acheterAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.FormValue("id") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required field: id")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	var vendeur int
	var etat string
	err = conn.QueryRow(context.Background(), "select vendeur, etat from annonce where annonce_id = $1", r.FormValue("id")).Scan(&vendeur, &etat)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Annonce not found"})
		return
	}

	if vendeur == token.ClientID {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot buy your own annonce")
		return
	}

	if etat != "D" {
		writeAPIError(w, http.StatusConflict, "invalid_request", "This annonce is no longer available")
		return
	}

	tag, err := conn.Exec(context.Background(), "update annonce set acheteur = $1 where annonce_id = $2 and etat = 'D'", token.ClientID, r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the annonce")
		return
	}

	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusConflict, "invalid_request", "This annonce is no longer available")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func supprimerAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	var vendeur int
	err = conn.QueryRow(context.Background(), "select vendeur from annonce where annonce_id = $1", r.FormValue("id")).Scan(&vendeur)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}

	isAdmin := token.Scope != nil && contains(token.Scope, "public:admin_annonces")
	if token.ClientID != vendeur && !isAdmin {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to delete this annonce")
		return
	}

	_, err = conn.Exec(context.Background(), "delete from annonce where annonce_id = $1", r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete the annonce")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.URL.Query().Get("from") != "" && r.URL.Query().Get("size") != "" {
		conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
			return
		}
		defer conn.Close(context.Background())

		listQuery := `select annonce.annonce_id, vendeur_client.nom, coalesce(acheteur_client.nom, ''), categorie_annonce.libelle,
			annonce.titre, annonce.prix, annonce.description, annonce.etat, annonce.taxe, annonce.image, annonce.barcode, annonce.date_publication
			from annonce
			inner join client vendeur_client on vendeur_client.client_id = annonce.vendeur
			left join client acheteur_client on acheteur_client.client_id = annonce.acheteur
			inner join categorie_annonce on categorie_annonce.categorie_id = annonce.categorie`
		listArgs := []any{}
		argPos := 1
		if r.URL.Query().Get("mes_achats") == "1" {
			listQuery += " where annonce.acheteur = $" + strconv.Itoa(argPos)
			listArgs = append(listArgs, token.ClientID)
			argPos++
		}
		listQuery += " order by annonce.date_publication desc limit $" + strconv.Itoa(argPos) + " offset $" + strconv.Itoa(argPos+1)
		listArgs = append(listArgs, r.URL.Query().Get("size"), r.URL.Query().Get("from"))

		rows, err := conn.Query(context.Background(), listQuery, listArgs...)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
			return
		}
		defer rows.Close()

		var annonces []Annonce

		for rows.Next() {
			var annonce Annonce
			err = rows.Scan(&annonce.AnnonceID, &annonce.Vendeur, &annonce.Acheteur, &annonce.Categorie, &annonce.Titre, &annonce.Prix, &annonce.Description, &annonce.Etat, &annonce.Taxe, &annonce.Image, &annonce.Barcode, &annonce.Date_Publication)
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the database")
				return
			}
			annonces = append(annonces, annonce)
		}

		jsonData, err := json.Marshal(annonces)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal the annonces")
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

		var annonce Annonce

		err = conn.QueryRow(context.Background(),
			`select annonce.annonce_id, vendeur_client.nom, coalesce(acheteur_client.nom, ''), categorie_annonce.libelle,
			annonce.titre, annonce.prix, annonce.description, annonce.etat, annonce.taxe, annonce.image, annonce.barcode, annonce.date_publication
			from annonce
			inner join client vendeur_client on vendeur_client.client_id = annonce.vendeur
			left join client acheteur_client on acheteur_client.client_id = annonce.acheteur
			inner join categorie_annonce on categorie_annonce.categorie_id = annonce.categorie
			where annonce.annonce_id = $1`,
			r.URL.Query().Get("id")).Scan(&annonce.AnnonceID, &annonce.Vendeur, &annonce.Acheteur, &annonce.Categorie, &annonce.Titre, &annonce.Prix, &annonce.Description, &annonce.Etat, &annonce.Taxe, &annonce.Image, &annonce.Barcode, &annonce.Date_Publication)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
			return
		}

		jsonData, err := json.Marshal(annonce)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal the annonce")
			return
		}

		var payload map[string]any
		if err := json.Unmarshal(jsonData, &payload); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to build the response")
			return
		}

		materiauxRows, err := conn.Query(context.Background(), "select materieau.materieau_id, materieau.nom from utilise inner join materieau on materieau.materieau_id = utilise.materieau_id where utilise.annonce_id = $1", annonce.AnnonceID)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the materiaux used")
			return
		}
		defer materiauxRows.Close()

		materiaux := make([]map[string]any, 0)
		for materiauxRows.Next() {
			var materieauID int
			var nom string
			if err := materiauxRows.Scan(&materieauID, &nom); err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the materiaux used")
				return
			}
			materiaux = append(materiaux, map[string]any{"materieau_id": materieauID, "nom": nom})
		}
		payload["materiaux"] = materiaux

		finalJSON, err := json.Marshal(payload)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal the annonce")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(finalJSON)
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "You must provide either the id of the annonce or the from and size query params.")
	}

	_ = token.ClientID
}

func getCategorieAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
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

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select categorie_id, libelle from categorie_annonce order by categorie_id limit $1 offset $2", size, from)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}
	defer rows.Close()

	categories := make([]CategorieAnnonce, 0)
	for rows.Next() {
		var categorie CategorieAnnonce
		err = rows.Scan(&categorie.CategorieID, &categorie.Libelle)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the database")
			return
		}
		categories = append(categories, categorie)
	}

	jsonData, err := json.Marshal(categories)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal categories")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

	_ = token.ClientID
}

func modifierAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	isAdmin := token.Scope != nil && contains(token.Scope, "public:admin_annonces")
	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	var vendeur int
	err = conn.QueryRow(context.Background(), "select vendeur from annonce where annonce_id = $1", r.FormValue("id")).Scan(&vendeur)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}

	if token.ClientID != vendeur && !isAdmin {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to modify this annonce")
		return
	}

	conn2, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn2.Close(context.Background())

	_, err = conn2.Exec(context.Background(), "update annonce set categorie = $1, titre = $2, prix = $3, description = $4, etat = $5, taxe = $6, image = $7 where annonce_id = $8", r.FormValue("categorie"), r.FormValue("titre"), r.FormValue("prix"), r.FormValue("description"), r.FormValue("etat"), r.FormValue("taxe"), r.FormValue("image"), r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the annonce")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func consulterMessagesAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "update message_annonce set statut = 'lu' where destinataire = $1 and statut = 'envoye'", token.ClientID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update messages status")
		return
	}

	rows, err := conn.Query(context.Background(), "select message_id, annonce_id, expediteur, destinataire, message, statut, date_envoi from message_annonce where expediteur = $1 or destinataire = $1 order by date_envoi desc", token.ClientID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}
	defer rows.Close()

	messages := make([]map[string]any, 0)
	for rows.Next() {
		var messageID int
		var annonceID int
		var expediteur string
		var destinataire string
		var message string
		var statut string
		var dateEnvoi string

		err = rows.Scan(&messageID, &annonceID, &expediteur, &destinataire, &message, &statut, &dateEnvoi)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the database")
			return
		}

		messages = append(messages, map[string]any{
			"message_id":   messageID,
			"annonce_id":   annonceID,
			"expediteur":   expediteur,
			"destinataire": destinataire,
			"message":      message,
			"statut":       statut,
			"date_envoi":   dateEnvoi,
		})
	}

	jsonPayload, err := json.Marshal(messages)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal messages")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonPayload)

	_ = r.URL.String()
}

func envoyerMessageAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	var vendeur int
	err = conn.QueryRow(context.Background(), "select vendeur from annonce where annonce_id = $1", r.FormValue("annonce_id")).Scan(&vendeur)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}

	isAdmin := token.Scope != nil && contains(token.Scope, "public:admin_annonces")

	if token.ClientID != vendeur && !isAdmin {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to send messages for this annonce")
		return
	}

	if r.FormValue("destinataire") == "" || r.FormValue("message") == "" || r.FormValue("annonce_id") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "annonce_id, destinataire and message are required")
		return
	}

	_, err = conn.Exec(context.Background(), "insert into message_annonce (annonce_id, expediteur, destinataire, message, statut) values ($1, $2, $3, $4, 'envoye')", r.FormValue("annonce_id"), token.ClientID, r.FormValue("destinataire"), r.FormValue("message"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to insert the message")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func supprimerMessageAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	var expediteur int
	err = conn.QueryRow(context.Background(), "select expediteur from message_annonce where message_id = $1", r.FormValue("id")).Scan(&expediteur)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
		return
	}

	isAdmin := token.Scope != nil && contains(token.Scope, "public:admin_annonces")

	if token.ClientID != expediteur && !isAdmin {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to delete this message")
		return
	}

	_, err = conn.Exec(context.Background(), "delete from message_annonce where message_id = $1", r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete the message")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func modifierMessageAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	writeAPIError(w, http.StatusForbidden, "forbidden", "Messages cannot be modified once sent. Status changes automatically from envoye to lu on GET by the receiver.")
	_ = token.ClientID
	_ = r.URL.String()
}



