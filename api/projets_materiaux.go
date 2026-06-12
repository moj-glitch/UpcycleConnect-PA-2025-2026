package main

import (
	"context"
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

func projetsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getProjet(w, r)
	case http.MethodPut:
		token := tryAuth(w, r)
		if !token.Active {
			return
		}
		deposerProjet(w, r, token)
	case http.MethodPatch:
		token := tryAuth(w, r)
		if !token.Active {
			return
		}
		modifierProjet(w, r, token)
	case http.MethodDelete:
		token := tryAuth(w, r)
		if !token.Active {
			return
		}
		supprimerProjet(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func categoriesProjetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}
	getCategorieProjet(w, r)
}

func autocompleteCategoriesProjetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}
	autocompleteCategorieProjet(w, r)
}

func materiauxHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getMateriau(w, r)
	case http.MethodPut:
		token := tryAuth(w, r)
		if !token.Active {
			return
		}
		deposerMateriau(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func hasProjectScope(token *IntrospectionPayload, scope string) bool {
	return token.Scope != nil && contains(token.Scope, scope)
}

func canManageProject(ctx context.Context, conn *pgx.Conn, token *IntrospectionPayload, entrepriseID int) bool {
	isAdmin := hasProjectScope(token, "pro:project_administrateur")
	isOwn := hasProjectScope(token, "pro:project_manager") && isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID)
	return isAdmin || isOwn
}

func parseProjectEntrepriseID(ctx context.Context, conn *pgx.Conn, r *http.Request, token *IntrospectionPayload) (int, error) {
	param := r.URL.Query().Get("entreprise_id")
	if param == "" {
		return getUserEntrepriseID(ctx, conn, token.ClientID)
	}
	return strconv.Atoi(param)
}

func saveProjetMedia(projectID int, timestamp int64, r *http.Request) (string, string, error) {
	if err := r.ParseMultipartForm(128 << 20); err != nil {
		return "", "", err
	}

	if err := os.MkdirAll(DATA_DIR, 0o755); err != nil {
		return "", "", err
	}

	imageFile, imageHeader, err := r.FormFile("image")
	if err != nil {
		return "", "", err
	}
	defer imageFile.Close()

	videoFile, videoHeader, err := r.FormFile("video")
	if err != nil {
		return "", "", err
	}
	defer videoFile.Close()

	imageExt := strings.ToLower(filepath.Ext(imageHeader.Filename))
	videoExt := strings.ToLower(filepath.Ext(videoHeader.Filename))
	if imageExt == "" {
		imageExt = ".bin"
	}
	if videoExt == "" {
		videoExt = ".bin"
	}

	imageName := fmt.Sprintf("%d_%d_img%s", projectID, timestamp, imageExt)
	videoName := fmt.Sprintf("%d_%d_vid%s", projectID, timestamp, videoExt)

	imagePath := filepath.Join(DATA_DIR, imageName)
	videoPath := filepath.Join(DATA_DIR, videoName)

	imgDst, err := os.Create(imagePath)
	if err != nil {
		return "", "", err
	}
	defer imgDst.Close()
	if _, err = io.Copy(imgDst, imageFile); err != nil {
		return "", "", err
	}

	vidDst, err := os.Create(videoPath)
	if err != nil {
		return "", "", err
	}
	defer vidDst.Close()
	if _, err = io.Copy(vidDst, videoFile); err != nil {
		return "", "", err
	}

	return imageName, videoName, nil
}

func deposerProjet(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	entrepriseID, err := parseProjectEntrepriseID(ctx, conn, r, token)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id or client not attached to an entreprise")
		return
	}

	if !canManageProject(ctx, conn, token, entrepriseID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot create projects for this entreprise")
		return
	}

	if err = r.ParseMultipartForm(128 << 20); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed multipart form")
		return
	}

	categorie := r.FormValue("categorie_projet")
	titre := r.FormValue("titre")
	texte := r.FormValue("texte")
	if categorie == "" || titre == "" || texte == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "categorie_projet, titre and texte are required")
		return
	}

	categorieID, err := strconv.Atoi(categorie)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "categorie_projet must be a valid integer")
		return
	}

	var projectID int
	err = conn.QueryRow(ctx, "insert into projet (entreprise_id, categorie_projet, titre, image, video, texte) values ($1, $2, $3, $4, $5, $6) returning projet_id", entrepriseID, categorieID, titre, "tmp", "tmp", texte).Scan(&projectID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create project")
		return
	}

	timestamp := time.Now().Unix()
	imageName, videoName, err := saveProjetMedia(projectID, timestamp, r)
	if err != nil {
		_, _ = conn.Exec(ctx, "delete from projet where projet_id=$1", projectID)
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "image and video files are required and must be readable")
		return
	}

	_, err = conn.Exec(ctx, "update projet set image=$1, video=$2 where projet_id=$3", imageName, videoName, projectID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update project media")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":   "Project created successfully",
		"projet_id": projectID,
		"image":     imageName,
		"video":     videoName,
	})
}

func getProjet(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		id, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}

		var p Projet
		err = conn.QueryRow(ctx, "select projet_id, entreprise_id, categorie_projet, titre, image, video, texte, date_creation from projet where projet_id=$1", id).Scan(&p.ProjetID, &p.EntrepriseID, &p.CategorieProjet, &p.Titre, &p.Image, &p.Video, &p.Texte, &p.DateCreation)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Project not found")
			return
		}
		writeJSON(w, http.StatusOK, p)
		return
	}

	fromParam := r.URL.Query().Get("from")
	sizeParam := r.URL.Query().Get("size")
	if fromParam == "" || sizeParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id or from and size are required")
		return
	}

	from, convErr := strconv.Atoi(fromParam)
	if convErr != nil || from < 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from must be >= 0")
		return
	}
	size, convErr := strconv.Atoi(sizeParam)
	if convErr != nil || size <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "size must be > 0")
		return
	}

	entrepriseParam := r.URL.Query().Get("entreprise_id")
	if entrepriseParam != "" {
		entrepriseID, convErr := strconv.Atoi(entrepriseParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "entreprise_id must be a valid integer")
			return
		}
		rows, err := conn.Query(ctx, "select projet_id, entreprise_id, categorie_projet, titre, image, video, texte, date_creation from projet where entreprise_id=$1 order by date_creation desc offset $2 limit $3", entrepriseID, from, size)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query projects")
			return
		}
		defer rows.Close()

		projects := make([]Projet, 0)
		for rows.Next() {
			var p Projet
			if err = rows.Scan(&p.ProjetID, &p.EntrepriseID, &p.CategorieProjet, &p.Titre, &p.Image, &p.Video, &p.Texte, &p.DateCreation); err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse projects")
				return
			}
			projects = append(projects, p)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"from":          from,
			"size":          size,
			"entreprise_id": entrepriseID,
			"total":         len(projects),
			"projets":       projects,
		})
		return
	}

	rows, err := conn.Query(ctx, "select projet_id, entreprise_id, categorie_projet, titre, image, video, texte, date_creation from projet order by date_creation desc offset $1 limit $2", from, size)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query projects")
		return
	}
	defer rows.Close()

	projects := make([]Projet, 0)
	for rows.Next() {
		var p Projet
		if err = rows.Scan(&p.ProjetID, &p.EntrepriseID, &p.CategorieProjet, &p.Titre, &p.Image, &p.Video, &p.Texte, &p.DateCreation); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse projects")
			return
		}
		projects = append(projects, p)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"from":    from,
		"size":    size,
		"total":   len(projects),
		"projets": projects,
	})
}

func modifierProjet(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}
	projectID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var entrepriseID int
	err = conn.QueryRow(ctx, "select entreprise_id from projet where projet_id=$1", projectID).Scan(&entrepriseID)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Project not found")
		return
	}

	if !canManageProject(ctx, conn, token, entrepriseID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot modify projects for this entreprise")
		return
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err = r.ParseMultipartForm(128 << 20); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed multipart form")
			return
		}
	} else {
		if err = r.ParseForm(); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
			return
		}
	}

	updates := make([]string, 0)
	args := make([]any, 0)
	argIndex := 1

	if value := r.FormValue("categorie_projet"); value != "" {
		categorieID, convErr := strconv.Atoi(value)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "categorie_projet must be a valid integer")
			return
		}
		updates = append(updates, "categorie_projet=$"+strconv.Itoa(argIndex))
		args = append(args, categorieID)
		argIndex++
	}
	if value := r.FormValue("titre"); value != "" {
		updates = append(updates, "titre=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}
	if value := r.FormValue("texte"); value != "" {
		updates = append(updates, "texte=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	timestamp := time.Now().Unix()
	if _, _, fileErr := r.FormFile("image"); fileErr == nil {
		imageFile, imageHeader, _ := r.FormFile("image")
		_ = imageFile.Close()
		ext := strings.ToLower(filepath.Ext(imageHeader.Filename))
		if ext == "" {
			ext = ".bin"
		}
		imageName := fmt.Sprintf("%d_%d_img%s", projectID, timestamp, ext)
		imgPath := filepath.Join(DATA_DIR, imageName)
		if err = os.MkdirAll(DATA_DIR, 0o755); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to prepare media directory")
			return
		}
		imgSrc, _, _ := r.FormFile("image")
		defer imgSrc.Close()
		imgDst, createErr := os.Create(imgPath)
		if createErr != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to save image")
			return
		}
		_, copyErr := io.Copy(imgDst, imgSrc)
		_ = imgDst.Close()
		if copyErr != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to save image")
			return
		}
		updates = append(updates, "image=$"+strconv.Itoa(argIndex))
		args = append(args, imageName)
		argIndex++
	}

	if _, _, fileErr := r.FormFile("video"); fileErr == nil {
		videoFile, videoHeader, _ := r.FormFile("video")
		_ = videoFile.Close()
		ext := strings.ToLower(filepath.Ext(videoHeader.Filename))
		if ext == "" {
			ext = ".bin"
		}
		videoName := fmt.Sprintf("%d_%d_vid%s", projectID, timestamp, ext)
		vidPath := filepath.Join(DATA_DIR, videoName)
		if err = os.MkdirAll(DATA_DIR, 0o755); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to prepare media directory")
			return
		}
		vidSrc, _, _ := r.FormFile("video")
		defer vidSrc.Close()
		vidDst, createErr := os.Create(vidPath)
		if createErr != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to save video")
			return
		}
		_, copyErr := io.Copy(vidDst, vidSrc)
		_ = vidDst.Close()
		if copyErr != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to save video")
			return
		}
		updates = append(updates, "video=$"+strconv.Itoa(argIndex))
		args = append(args, videoName)
		argIndex++
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, projectID)
	query := "update projet set " + strings.Join(updates, ", ") + " where projet_id=$" + strconv.Itoa(argIndex)
	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update project")
		return
	}
	if result.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Project not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Project updated successfully", "projet_id": projectID})
}

func supprimerProjet(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}
	projectID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var entrepriseID int
	err = conn.QueryRow(ctx, "select entreprise_id from projet where projet_id=$1", projectID).Scan(&entrepriseID)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Project not found")
		return
	}

	if !canManageProject(ctx, conn, token, entrepriseID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot delete projects for this entreprise")
		return
	}

	result, err := conn.Exec(ctx, "delete from projet where projet_id=$1", projectID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete project")
		return
	}
	if result.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Project not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Project deleted successfully", "projet_id": projectID})
}

func getCategorieProjet(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		id, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}
		var c CategorieProjet
		err = conn.QueryRow(ctx, "select categorie_projet_id, libelle from categorie_projet where categorie_projet_id=$1", id).Scan(&c.CategorieProjetID, &c.Libelle)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Category not found")
			return
		}
		writeJSON(w, http.StatusOK, c)
		return
	}

	fromParam := r.URL.Query().Get("from")
	sizeParam := r.URL.Query().Get("size")
	if fromParam == "" || sizeParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id or from and size are required")
		return
	}

	from, convErr := strconv.Atoi(fromParam)
	if convErr != nil || from < 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from must be >= 0")
		return
	}
	size, convErr := strconv.Atoi(sizeParam)
	if convErr != nil || size <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "size must be > 0")
		return
	}

	rows, err := conn.Query(ctx, "select categorie_projet_id, libelle from categorie_projet order by libelle asc offset $1 limit $2", from, size)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query categories")
		return
	}
	defer rows.Close()

	categories := make([]CategorieProjet, 0)
	for rows.Next() {
		var c CategorieProjet
		if err = rows.Scan(&c.CategorieProjetID, &c.Libelle); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse categories")
			return
		}
		categories = append(categories, c)
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(categories), "categories": categories})
}

func autocompleteCategorieProjet(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	size := 10
	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		parsedSize, convErr := strconv.Atoi(sizeParam)
		if convErr != nil || parsedSize <= 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "size must be a positive integer")
			return
		}
		size = parsedSize
	}
	if size > 50 {
		size = 50
	}

	pattern := "%" + q + "%"
	rows, err := conn.Query(ctx, "select categorie_projet_id, libelle from categorie_projet where libelle ilike $1 order by libelle asc limit $2", pattern, size)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query categories")
		return
	}
	defer rows.Close()

	categories := make([]CategorieProjet, 0)
	for rows.Next() {
		var c CategorieProjet
		if err = rows.Scan(&c.CategorieProjetID, &c.Libelle); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse categories")
			return
		}
		categories = append(categories, c)
	}

	writeJSON(w, http.StatusOK, map[string]any{"query": q, "size": size, "total": len(categories), "categories": categories})
}

func deposerMateriau(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	nom := r.FormValue("nom")
	if nom == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "nom is required")
		return
	}

	getFloat := func(key string) (float64, error) {
		value := r.FormValue(key)
		if value == "" {
			return 0, fmt.Errorf("missing %s", key)
		}
		return strconv.ParseFloat(value, 64)
	}

	densite, err := getFloat("densite")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "densite is required")
		return
	}
	prixKG, err := getFloat("prix_kg")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "prix_kg is required")
		return
	}
	durete, err := getFloat("durete")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "durete is required")
		return
	}
	fonteDegree, err := getFloat("fonte_degree")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "fonte_degree is required")
		return
	}
	elasticiteYoung, err := getFloat("elasticite_young")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "elasticite_young is required")
		return
	}
	resTorsion, err := getFloat("resistance_torsion")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "resistance_torsion is required")
		return
	}
	resCompression, err := getFloat("resistance_compression")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "resistance_compression is required")
		return
	}
	resilience, err := getFloat("resilience")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "resilience is required")
		return
	}
	condThermique, err := getFloat("conductivite_thermique")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "conductivite_thermique is required")
		return
	}
	condElectrique, err := getFloat("conductivite_electrique")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "conductivite_electrique is required")
		return
	}
	dilatThermique, err := getFloat("dilatation_thermique")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "dilatation_thermique is required")
		return
	}
	porosite, err := getFloat("porosite")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "porosite is required")
		return
	}
	opacite, err := getFloat("opacite")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "opacite is required")
		return
	}
	magnetisme, err := getFloat("magnetisme")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "magnetisme is required")
		return
	}
	accoustique, err := getFloat("accoustique")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "accoustique is required")
		return
	}
	empreinteCO2, err := getFloat("empreinte_co2")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "empreinte_co2 is required")
		return
	}
	toxicite, err := getFloat("toxicite")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "toxicite is required")
		return
	}

	recyclable := r.FormValue("recyclable")
	if recyclable == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "recyclable is required")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "insert into materieau (nom, densite, prix_kg, durete, fonte_degree, elasticite_young, resistance_torsion, resistance_compression, resilience, conductivite_thermique, conductivite_electrique, dilatation_thermique, porosite, opacite, magnetisme, accoustique, recyclable, empreinte_co2, toxicite) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)", nom, densite, prixKG, durete, fonteDegree, elasticiteYoung, resTorsion, resCompression, resilience, condThermique, condElectrique, dilatThermique, porosite, opacite, magnetisme, accoustique, recyclable, empreinteCO2, toxicite)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create materiau")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Materiau created successfully"})
}

func getMateriau(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		id, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}
		var m Materiau
		err = conn.QueryRow(ctx, "select materieau_id, nom, densite, prix_kg, durete, fonte_degree, elasticite_young, resistance_torsion, resistance_compression, resilience, conductivite_thermique, conductivite_electrique, dilatation_thermique, porosite, opacite, magnetisme, accoustique, recyclable, empreinte_co2, toxicite from materieau where materieau_id=$1", id).Scan(&m.MaterieauID, &m.Nom, &m.Densite, &m.PrixKG, &m.Durete, &m.FonteDegree, &m.ElasticiteYoung, &m.ResistanceTorsion, &m.ResistanceCompression, &m.Resilience, &m.ConductiviteThermique, &m.ConductiviteElectrique, &m.DilatationThermique, &m.Porosite, &m.Opacite, &m.Magnetisme, &m.Accoustique, &m.Recyclable, &m.EmpreinteCO2, &m.Toxicite)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Materiau not found")
			return
		}
		writeJSON(w, http.StatusOK, m)
		return
	}

	filterClauses := make([]string, 0)
	args := make([]any, 0)
	argIndex := 1
	filterCount := 0

	addFloatFilter := func(param string, clausePrefix string) {
		value := r.URL.Query().Get(param)
		if value == "" {
			return
		}
		parsed, convErr := strconv.ParseFloat(value, 64)
		if convErr != nil {
			filterClauses = append(filterClauses, "INVALID")
			return
		}
		filterClauses = append(filterClauses, clausePrefix+"$"+strconv.Itoa(argIndex))
		args = append(args, parsed)
		argIndex++
		filterCount++
	}

	addFloatFilter("densite_min", "densite >= ")
	addFloatFilter("densite_max", "densite <= ")
	addFloatFilter("prix_kg_min", "prix_kg >= ")
	addFloatFilter("prix_kg_max", "prix_kg <= ")
	addFloatFilter("durete_min", "durete >= ")
	addFloatFilter("durete_max", "durete <= ")
	addFloatFilter("toxicite_max", "toxicite <= ")
	addFloatFilter("empreinte_co2_max", "empreinte_co2 <= ")

	recyclable := r.URL.Query().Get("recyclable")
	if recyclable != "" {
		filterClauses = append(filterClauses, "recyclable = $"+strconv.Itoa(argIndex))
		args = append(args, recyclable)
		argIndex++
		filterCount++
	}

	for _, c := range filterClauses {
		if c == "INVALID" {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "One or more filter values are invalid")
			return
		}
	}

	if filterCount < 2 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "At least two filter criteria are required")
		return
	}

	baseQuery := "select materieau_id, nom, densite, prix_kg, durete, fonte_degree, elasticite_young, resistance_torsion, resistance_compression, resilience, conductivite_thermique, conductivite_electrique, dilatation_thermique, porosite, opacite, magnetisme, accoustique, recyclable, empreinte_co2, toxicite from materieau where " + strings.Join(filterClauses, " and ")

	idsParam := strings.TrimSpace(r.URL.Query().Get("ids"))
	if idsParam != "" {
		idParts := strings.Split(idsParam, ",")
		ids := make([]int, 0, len(idParts))
		for _, part := range idParts {
			p := strings.TrimSpace(part)
			if p == "" {
				continue
			}
			parsedID, convErr := strconv.Atoi(p)
			if convErr != nil {
				writeAPIError(w, http.StatusBadRequest, "invalid_request", "ids must contain integers")
				return
			}
			ids = append(ids, parsedID)
		}
		if len(ids) == 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "ids list cannot be empty")
			return
		}

		idPlaceholders := make([]string, 0, len(ids))
		for _, id := range ids {
			idPlaceholders = append(idPlaceholders, "$"+strconv.Itoa(argIndex))
			args = append(args, id)
			argIndex++
		}

		query := baseQuery + " and materieau_id in (" + strings.Join(idPlaceholders, ",") + ") order by nom asc"
		rows, queryErr := conn.Query(ctx, query, args...)
		if queryErr != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query materiaux")
			return
		}
		defer rows.Close()

		list := make([]Materiau, 0)
		for rows.Next() {
			var m Materiau
			if err = rows.Scan(&m.MaterieauID, &m.Nom, &m.Densite, &m.PrixKG, &m.Durete, &m.FonteDegree, &m.ElasticiteYoung, &m.ResistanceTorsion, &m.ResistanceCompression, &m.Resilience, &m.ConductiviteThermique, &m.ConductiviteElectrique, &m.DilatationThermique, &m.Porosite, &m.Opacite, &m.Magnetisme, &m.Accoustique, &m.Recyclable, &m.EmpreinteCO2, &m.Toxicite); err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse materiaux")
				return
			}
			list = append(list, m)
		}

		writeJSON(w, http.StatusOK, map[string]any{"mode": "list", "total": len(list), "materiaux": list})
		return
	}

	fromParam := r.URL.Query().Get("from")
	sizeParam := r.URL.Query().Get("size")
	if fromParam == "" || sizeParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Use either ids list or pagination (from and size)")
		return
	}

	from, convErr := strconv.Atoi(fromParam)
	if convErr != nil || from < 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from must be >= 0")
		return
	}
	size, convErr := strconv.Atoi(sizeParam)
	if convErr != nil || size <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "size must be > 0")
		return
	}

	query := baseQuery + " order by nom asc offset $" + strconv.Itoa(argIndex) + " limit $" + strconv.Itoa(argIndex+1)
	args = append(args, from, size)
	rows, queryErr := conn.Query(ctx, query, args...)
	if queryErr != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query materiaux")
		return
	}
	defer rows.Close()

	list := make([]Materiau, 0)
	for rows.Next() {
		var m Materiau
		if err = rows.Scan(&m.MaterieauID, &m.Nom, &m.Densite, &m.PrixKG, &m.Durete, &m.FonteDegree, &m.ElasticiteYoung, &m.ResistanceTorsion, &m.ResistanceCompression, &m.Resilience, &m.ConductiviteThermique, &m.ConductiviteElectrique, &m.DilatationThermique, &m.Porosite, &m.Opacite, &m.Magnetisme, &m.Accoustique, &m.Recyclable, &m.EmpreinteCO2, &m.Toxicite); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse materiaux")
			return
		}
		list = append(list, m)
	}

	writeJSON(w, http.StatusOK, map[string]any{"mode": "pagination", "from": from, "size": size, "total": len(list), "materiaux": list})
}
