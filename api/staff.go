package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func evenementsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodPut:
			deposerEvenement(w, r, token)
		case http.MethodGet:
			getEvenement(w, r, token)
		case http.MethodPatch:
			modifierEvenement(w, r, token)
		case http.MethodDelete:
			supprimerEvenement(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func tutorielsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)

	if token.Active {
		switch r.Method {
		case http.MethodPut:
			deposerTutoriel(w, r, token)
		case http.MethodGet:
			getTutoriel(w, r, token)
		case http.MethodPatch:
			modifierTutoriel(w, r, token)
		case http.MethodDelete:
			supprimerTutoriel(w, r, token)
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		}
	}
}

func hasScope(token *IntrospectionPayload, scope string) bool {
	return token.Scope != nil && contains(token.Scope, scope)
}

func deposerEvenement(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !hasScope(token, "evenements:manager") {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to create events")
		return
	}

	if r.FormValue("date") == "" || r.FormValue("nom") == "" || r.FormValue("description") == "" || r.FormValue("statut") == "" || r.FormValue("categorie") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "date, nom, description, statut and categorie are required")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "insert into evenement (date, nom, description, statut, categorie) values ($1, $2, $3, $4, $5)", r.FormValue("date"), r.FormValue("nom"), r.FormValue("description"), r.FormValue("statut"), r.FormValue("categorie"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to insert event")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getEvenement(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.URL.Query().Get("from") != "" && r.URL.Query().Get("size") != "" {
		conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
			return
		}
		defer conn.Close(context.Background())

		rows, err := conn.Query(context.Background(), "select evenement_id, date, nom, description, statut, categorie from evenement order by date desc limit $1 offset $2", r.URL.Query().Get("size"), r.URL.Query().Get("from"))
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
			return
		}
		defer rows.Close()

		evenements := make([]Evenement, 0)
		for rows.Next() {
			var evenement Evenement
			err = rows.Scan(&evenement.EvenementID, &evenement.Date, &evenement.Nom, &evenement.Description, &evenement.Statut, &evenement.Categorie)
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the database")
				return
			}
			evenements = append(evenements, evenement)
		}

		jsonData, err := json.Marshal(evenements)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal events")
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

		var evenement Evenement
		err = conn.QueryRow(context.Background(), "select evenement_id, date, nom, description, statut, categorie from evenement where evenement_id = $1", r.URL.Query().Get("id")).Scan(&evenement.EvenementID, &evenement.Date, &evenement.Nom, &evenement.Description, &evenement.Statut, &evenement.Categorie)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
			return
		}

		jsonData, err := json.Marshal(evenement)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal event")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "You must provide either the id of the event or the from and size query params.")
	}

	_ = token.ClientID
}

func modifierEvenement(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !hasScope(token, "evenements:manager") {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to modify events")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "update evenement set date = $1, nom = $2, description = $3, statut = $4, categorie = $5 where evenement_id = $6", r.FormValue("date"), r.FormValue("nom"), r.FormValue("description"), r.FormValue("statut"), r.FormValue("categorie"), r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update event")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func supprimerEvenement(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !hasScope(token, "evenements:manager") {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to delete events")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "delete from evenement where evenement_id = $1", r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete event")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func saveTutorielVideo(r *http.Request, token *IntrospectionPayload) (string, error) {
	err := r.ParseMultipartForm(128 << 20)
	if err != nil {
		return "", err
	}

	videoFile, videoHeader, err := r.FormFile("video")
	if err != nil {
		return "", err
	}
	defer videoFile.Close()

	err = os.MkdirAll(DATA_DIR, 0o755)
	if err != nil {
		return "", err
	}

	uploaderName := token.Username
	if uploaderName == "" {
		uploaderName = strconv.Itoa(token.ClientID)
	}
	uploaderName = strings.ReplaceAll(uploaderName, " ", "_")
	uploaderName = strings.ReplaceAll(uploaderName, "/", "_")
	uploaderName = strings.ReplaceAll(uploaderName, "\\", "_")

	ext := filepath.Ext(videoHeader.Filename)
	if ext == "" {
		ext = ".bin"
	}

	videoName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uploaderName, ext)
	videoPath := filepath.Join(DATA_DIR, videoName)

	dst, err := os.Create(videoPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, videoFile)
	if err != nil {
		return "", err
	}

	return videoName, nil
}

func deposerTutoriel(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !hasScope(token, "tutorials:content_manager") {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to create tutorials")
		return
	}

	videoName, err := saveTutorielVideo(r, token)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "video file is required and must be readable")
		return
	}

	if r.FormValue("titre") == "" || r.FormValue("article") == "" || r.FormValue("categorie") == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "titre, article and categorie are required")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "insert into tutoriel (titre, article, video, categorie) values ($1, $2, $3, $4)", r.FormValue("titre"), r.FormValue("article"), videoName, r.FormValue("categorie"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to insert tutorial")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getTutoriel(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if r.URL.Query().Get("from") != "" && r.URL.Query().Get("size") != "" {
		conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
			return
		}
		defer conn.Close(context.Background())

		rows, err := conn.Query(context.Background(), "select tutoriel_id, titre, date_creation, article, video, categorie from tutoriel order by date_creation desc limit $1 offset $2", r.URL.Query().Get("size"), r.URL.Query().Get("from"))
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
			return
		}
		defer rows.Close()

		tutoriels := make([]Tutoriel, 0)
		for rows.Next() {
			var tutoriel Tutoriel
			err = rows.Scan(&tutoriel.TutorielID, &tutoriel.Titre, &tutoriel.DateCreation, &tutoriel.Article, &tutoriel.Video, &tutoriel.Categorie)
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the database")
				return
			}
			tutoriels = append(tutoriels, tutoriel)
		}

		jsonData, err := json.Marshal(tutoriels)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal tutorials")
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

		var tutoriel Tutoriel
		err = conn.QueryRow(context.Background(), "select tutoriel_id, titre, date_creation, article, video, categorie from tutoriel where tutoriel_id = $1", r.URL.Query().Get("id")).Scan(&tutoriel.TutorielID, &tutoriel.Titre, &tutoriel.DateCreation, &tutoriel.Article, &tutoriel.Video, &tutoriel.Categorie)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the database")
			return
		}

		jsonData, err := json.Marshal(tutoriel)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to marshal tutorial")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "You must provide either the id of the tutorial or the from and size query params.")
	}

	_ = token.ClientID
}

func modifierTutoriel(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !hasScope(token, "tutorials:content_manager") {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to modify tutorials")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	video := r.FormValue("video")
	if video == "" {
		video = r.FormValue("existing_video")
	}

	_, err = conn.Exec(context.Background(), "update tutoriel set titre = $1, article = $2, video = $3, categorie = $4 where tutoriel_id = $5", r.FormValue("titre"), r.FormValue("article"), video, r.FormValue("categorie"), r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update tutorial")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func supprimerTutoriel(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !hasScope(token, "tutorials:content_manager") {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "You don't have the right to delete tutorials")
		return
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "delete from tutoriel where tutoriel_id = $1", r.FormValue("id"))
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete tutorial")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



