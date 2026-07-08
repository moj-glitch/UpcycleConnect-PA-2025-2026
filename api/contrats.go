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

func contratsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		deposerContrat(w, r, token)
	case http.MethodGet:
		getContrat(w, r, token)
	case http.MethodPatch:
		modifierContrat(w, r, token)
	case http.MethodDelete:
		supprimerContrat(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func hasContratScope(token *IntrospectionPayload, scope string) bool {
	return token.Scope != nil && contains(token.Scope, scope)
}

func hasAnyContratManagementScope(token *IntrospectionPayload) bool {
	return hasContratScope(token, "pro:gestionnaire_contrats") || hasContratScope(token, "pro:contrat_administrateur")
}

func getUserEntrepriseID(ctx context.Context, conn *pgx.Conn, clientID int) (int, error) {
	var entrepriseID int
	err := conn.QueryRow(ctx, "select entreprise_id from employe where client_id=$1 order by entreprise_id asc limit 1", clientID).Scan(&entrepriseID)
	return entrepriseID, err
}

func parseEntrepriseIDFromRequest(ctx context.Context, conn *pgx.Conn, r *http.Request, token *IntrospectionPayload) (int, error) {
	param := r.URL.Query().Get("entreprise_id")
	if param == "" {
		return getUserEntrepriseID(ctx, conn, token.ClientID)
	}
	return strconv.Atoi(param)
}

func isContratManagerForEntreprise(ctx context.Context, conn *pgx.Conn, token *IntrospectionPayload, entrepriseID int) bool {
	isSupport := hasContratScope(token, "pro:contrat_administrateur")
	isOwnManager := hasContratScope(token, "pro:gestionnaire_contrats") && isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID)
	return isSupport || isOwnManager
}

func canManageContractDomain(ctx context.Context, conn *pgx.Conn, r *http.Request, token *IntrospectionPayload) bool {
	if hasContratScope(token, "pro:contrat_administrateur") {
		return true
	}
	if !hasContratScope(token, "pro:gestionnaire_contrats") {
		return false
	}
	entrepriseID, err := parseEntrepriseIDFromRequest(ctx, conn, r, token)
	if err != nil {
		return false
	}
	return isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID)
}

func saveContratPDF(r *http.Request, entrepriseID int) (string, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return "", err
	}

	pdfFile, pdfHeader, err := r.FormFile("pdf")
	if err != nil {
		return "", err
	}
	defer pdfFile.Close()

	if err := os.MkdirAll(DATA_DIR, 0o755); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(pdfHeader.Filename))
	if ext == "" {
		ext = ".pdf"
	}

	pdfName := fmt.Sprintf("%d_%d_contrat%s", entrepriseID, time.Now().Unix(), ext)
	pdfPath := filepath.Join(DATA_DIR, pdfName)

	dst, err := os.Create(pdfPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, pdfFile); err != nil {
		return "", err
	}

	return pdfName, nil
}

func deposerContrat(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	entrepriseID, err := parseEntrepriseIDFromRequest(ctx, conn, r, token)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id or client not attached to an entreprise")
		return
	}

	if !isContratManagerForEntreprise(ctx, conn, token, entrepriseID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot create contracts for this entreprise")
		return
	}

	pdfName, err := saveContratPDF(r, entrepriseID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "pdf file is required and must be readable")
		return
	}

	categorieContrat := r.FormValue("categorie_contrat_id")
	description := r.FormValue("description")
	depense := r.FormValue("depense")
	gain := r.FormValue("gain")
	dateFin := r.FormValue("date_fin")

	if categorieContrat == "" || description == "" || depense == "" || gain == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "categorie_contrat_id, description, depense and gain are required")
		return
	}

	depenseValue, err := strconv.ParseFloat(depense, 64)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "depense must be a number")
		return
	}

	gainValue, err := strconv.ParseFloat(gain, 64)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "gain must be a number")
		return
	}

	categorieID, err := strconv.Atoi(categorieContrat)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "categorie_contrat_id must be a valid integer")
		return
	}

	var newContratID int
	if dateFin == "" {
		err = conn.QueryRow(
			ctx,
			"insert into contrat (entreprise_id, categorie_contrat_id, description, depense, gain, pdf) values ($1, $2, $3, $4, $5, $6) returning contrat_id",
			entrepriseID, categorieID, description, depenseValue, gainValue, pdfName,
		).Scan(&newContratID)
	} else {
		err = conn.QueryRow(
			ctx,
			"insert into contrat (entreprise_id, categorie_contrat_id, description, depense, gain, pdf, date_fin) values ($1, $2, $3, $4, $5, $6, $7) returning contrat_id",
			entrepriseID, categorieID, description, depenseValue, gainValue, pdfName, dateFin,
		).Scan(&newContratID)
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to insert contrat")
		return
	}

	for _, tiersID := range r.Form["tiers"] {
		if tiersID == "" {
			continue
		}
		_, err = conn.Exec(ctx, "insert into actor (tiers_id, contrat_id) values ($1, $2) on conflict do nothing", tiersID, newContratID)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to link a tiers to the contrat")
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":       "Contrat created successfully",
		"contrat_id":    newContratID,
		"entreprise_id": entrepriseID,
		"pdf":           pdfName,
	})
}

func getContrat(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		contratID, err := strconv.Atoi(idParam)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}

		var contrat Contrat
		err = conn.QueryRow(
			ctx,
			"select contrat_id, entreprise_id, categorie_contrat_id, description, depense, gain, pdf, date_debut, date_fin from contrat where contrat_id=$1",
			contratID,
		).Scan(&contrat.ContratID, &contrat.EntrepriseID, &contrat.CategorieContratID, &contrat.Description, &contrat.Depense, &contrat.Gain, &contrat.PDF, &contrat.DateDebut, &contrat.DateFin)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Contrat not found")
			return
		}

		if !isContratManagerForEntreprise(ctx, conn, token, contrat.EntrepriseID) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot access this contrat")
			return
		}

		tiersRows, err := conn.Query(ctx, "select tiers.tiers_id, tiers.nom from actor inner join tiers on tiers.tiers_id = actor.tiers_id where actor.contrat_id = $1 order by tiers.nom asc", contratID)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the contrat tiers")
			return
		}
		defer tiersRows.Close()

		tiersList := make([]map[string]any, 0)
		for tiersRows.Next() {
			var tiersID int
			var nom string
			if err := tiersRows.Scan(&tiersID, &nom); err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the contrat tiers")
				return
			}
			tiersList = append(tiersList, map[string]any{"tiers_id": tiersID, "nom": nom})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"contrat_id":           contrat.ContratID,
			"entreprise_id":        contrat.EntrepriseID,
			"categorie_contrat_id": contrat.CategorieContratID,
			"description":          contrat.Description,
			"depense":              contrat.Depense,
			"gain":                 contrat.Gain,
			"pdf":                  contrat.PDF,
			"date_debut":           contrat.DateDebut,
			"date_fin":             contrat.DateFin,
			"tiers":                tiersList,
		})
		return
	}

	fromParam := r.URL.Query().Get("from")
	sizeParam := r.URL.Query().Get("size")
	if fromParam == "" || sizeParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from and size are required")
		return
	}

	from, err := strconv.Atoi(fromParam)
	if err != nil || from < 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from must be >= 0")
		return
	}

	size, err := strconv.Atoi(sizeParam)
	if err != nil || size <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "size must be > 0")
		return
	}

	entrepriseID, err := parseEntrepriseIDFromRequest(ctx, conn, r, token)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id or client not attached to an entreprise")
		return
	}

	if !isContratManagerForEntreprise(ctx, conn, token, entrepriseID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot access contrats for this entreprise")
		return
	}

	query := "select contrat.contrat_id, contrat.entreprise_id, contrat.categorie_contrat_id, contrat.description, contrat.depense, contrat.gain, contrat.pdf, contrat.date_debut, contrat.date_fin from contrat"
	args := []any{entrepriseID}
	argPos := 2
	conditions := " where contrat.entreprise_id=$1"

	if tiersQ := r.URL.Query().Get("tiers"); tiersQ != "" {
		query += " inner join actor on actor.contrat_id = contrat.contrat_id inner join tiers on tiers.tiers_id = actor.tiers_id"
		conditions += " and tiers.nom ilike $" + strconv.Itoa(argPos)
		args = append(args, "%"+tiersQ+"%")
		argPos++
	}

	if q := r.URL.Query().Get("q"); q != "" {
		conditions += " and contrat.description ilike $" + strconv.Itoa(argPos)
		args = append(args, "%"+q+"%")
		argPos++
	}

	query += conditions + " order by contrat.date_debut desc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query contrats")
		return
	}
	defer rows.Close()

	contrats := make([]Contrat, 0)
	for rows.Next() {
		var contrat Contrat
		err = rows.Scan(&contrat.ContratID, &contrat.EntrepriseID, &contrat.CategorieContratID, &contrat.Description, &contrat.Depense, &contrat.Gain, &contrat.PDF, &contrat.DateDebut, &contrat.DateFin)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse contrats")
			return
		}
		contrats = append(contrats, contrat)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"from":          from,
		"size":          size,
		"entreprise_id": entrepriseID,
		"total":         len(contrats),
		"contrats":      contrats,
	})
}

func modifierContrat(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	contratID, err := strconv.Atoi(idParam)
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
	err = conn.QueryRow(ctx, "select entreprise_id from contrat where contrat_id=$1", contratID).Scan(&entrepriseID)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Contrat not found")
		return
	}

	if !isContratManagerForEntreprise(ctx, conn, token, entrepriseID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot modify this contrat")
		return
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	pdfName := ""
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err = r.ParseMultipartForm(32 << 20); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed multipart form")
			return
		}
		if _, _, fileErr := r.FormFile("pdf"); fileErr == nil {
			pdfName, err = saveContratPDF(r, entrepriseID)
			if err != nil {
				writeAPIError(w, http.StatusBadRequest, "invalid_request", "pdf file is not readable")
				return
			}
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

	if value := r.FormValue("categorie_contrat_id"); value != "" {
		categorieID, convErr := strconv.Atoi(value)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "categorie_contrat_id must be a valid integer")
			return
		}
		updates = append(updates, "categorie_contrat_id=$"+strconv.Itoa(argIndex))
		args = append(args, categorieID)
		argIndex++
	}

	if value := r.FormValue("description"); value != "" {
		updates = append(updates, "description=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	if value := r.FormValue("depense"); value != "" {
		depenseValue, convErr := strconv.ParseFloat(value, 64)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "depense must be a number")
			return
		}
		updates = append(updates, "depense=$"+strconv.Itoa(argIndex))
		args = append(args, depenseValue)
		argIndex++
	}

	if value := r.FormValue("gain"); value != "" {
		gainValue, convErr := strconv.ParseFloat(value, 64)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "gain must be a number")
			return
		}
		updates = append(updates, "gain=$"+strconv.Itoa(argIndex))
		args = append(args, gainValue)
		argIndex++
	}

	if value := r.FormValue("date_fin"); value != "" {
		updates = append(updates, "date_fin=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	if pdfName != "" {
		updates = append(updates, "pdf=$"+strconv.Itoa(argIndex))
		args = append(args, pdfName)
		argIndex++
	}

	tiersProvided := len(r.Form["tiers"]) > 0

	if len(updates) == 0 && !tiersProvided {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	if len(updates) > 0 {
		args = append(args, contratID)
		query := "update contrat set " + strings.Join(updates, ", ") + " where contrat_id=$" + strconv.Itoa(argIndex)
		_, err = conn.Exec(ctx, query, args...)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update contrat")
			return
		}
	}

	if tiersProvided {
		_, err = conn.Exec(ctx, "delete from actor where contrat_id = $1", contratID)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to reset the contrat tiers")
			return
		}
		for _, tiersID := range r.Form["tiers"] {
			if tiersID == "" {
				continue
			}
			_, err = conn.Exec(ctx, "insert into actor (tiers_id, contrat_id) values ($1, $2) on conflict do nothing", tiersID, contratID)
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to link a tiers to the contrat")
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":    "Contrat updated successfully",
		"contrat_id": contratID,
	})
}

func supprimerContrat(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	contratID, err := strconv.Atoi(idParam)
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
	err = conn.QueryRow(ctx, "select entreprise_id from contrat where contrat_id=$1", contratID).Scan(&entrepriseID)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Contrat not found")
		return
	}

	if !isContratManagerForEntreprise(ctx, conn, token, entrepriseID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot delete this contrat")
		return
	}

	result, err := conn.Exec(ctx, "delete from contrat where contrat_id=$1", contratID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete contrat")
		return
	}

	if result.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Contrat not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":    "Contrat deleted successfully",
		"contrat_id": contratID,
	})
}

func tiersHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		deposerTier(w, r, token)
	case http.MethodGet:
		getTier(w, r, token)
	case http.MethodPatch:
		modifierTier(w, r, token)
	case http.MethodDelete:
		supprimerTier(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func autocompleteTiersHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if !canManageContractDomain(ctx, conn, r, token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot access tiers")
		return
	}

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
	rows, err := conn.Query(
		ctx,
		"select tiers_id, nom, adresse, code_postal, ville, siret from tiers where nom ilike $1 or coalesce(siret,'') ilike $1 order by nom asc limit $2",
		pattern, size,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query tiers")
		return
	}
	defer rows.Close()

	tiers := make([]Tier, 0)
	for rows.Next() {
		var t Tier
		if err = rows.Scan(&t.TiersID, &t.Nom, &t.Adresse, &t.CodePostal, &t.Ville, &t.Siret); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse tiers")
			return
		}
		tiers = append(tiers, t)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"query": q,
		"size":  size,
		"total": len(tiers),
		"tiers": tiers,
	})
}

func deposerTier(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if !canManageContractDomain(ctx, conn, r, token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot create tiers")
		return
	}

	nom := r.FormValue("nom")
	adresse := r.FormValue("adresse")
	codePostal := r.FormValue("code_postal")
	ville := r.FormValue("ville")
	siret := strings.TrimSpace(r.FormValue("siret"))

	if nom == "" || adresse == "" || codePostal == "" || ville == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "nom, adresse, code_postal and ville are required")
		return
	}

	var newID int
	if siret == "" {
		err = conn.QueryRow(ctx, "insert into tiers (nom, adresse, code_postal, ville, siret) values ($1, $2, $3, $4, null) returning tiers_id", nom, adresse, codePostal, ville).Scan(&newID)
	} else {
		err = conn.QueryRow(ctx, "insert into tiers (nom, adresse, code_postal, ville, siret) values ($1, $2, $3, $4, $5) returning tiers_id", nom, adresse, codePostal, ville, siret).Scan(&newID)
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create tier")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "Tier created successfully",
		"tiers_id": newID,
	})
}

func getTier(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if !canManageContractDomain(ctx, conn, r, token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot access tiers")
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		tiersID, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}

		var t Tier
		err = conn.QueryRow(ctx, "select tiers_id, nom, adresse, code_postal, ville, siret from tiers where tiers_id=$1", tiersID).Scan(&t.TiersID, &t.Nom, &t.Adresse, &t.CodePostal, &t.Ville, &t.Siret)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Tier not found")
			return
		}

		writeJSON(w, http.StatusOK, t)
		return
	}

	fromParam := r.URL.Query().Get("from")
	sizeParam := r.URL.Query().Get("size")
	if fromParam == "" || sizeParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from and size are required")
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

	rows, err := conn.Query(ctx, "select tiers_id, nom, adresse, code_postal, ville, siret from tiers order by nom asc offset $1 limit $2", from, size)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query tiers")
		return
	}
	defer rows.Close()

	tiers := make([]Tier, 0)
	for rows.Next() {
		var t Tier
		if err = rows.Scan(&t.TiersID, &t.Nom, &t.Adresse, &t.CodePostal, &t.Ville, &t.Siret); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse tiers")
			return
		}
		tiers = append(tiers, t)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"from":  from,
		"size":  size,
		"total": len(tiers),
		"tiers": tiers,
	})
}

func modifierTier(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	tiersID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
		return
	}

	if err = r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if !canManageContractDomain(ctx, conn, r, token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot modify tiers")
		return
	}

	updates := make([]string, 0)
	args := make([]any, 0)
	argIndex := 1

	if value := r.FormValue("nom"); value != "" {
		updates = append(updates, "nom=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}
	if value := r.FormValue("adresse"); value != "" {
		updates = append(updates, "adresse=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}
	if value := r.FormValue("code_postal"); value != "" {
		updates = append(updates, "code_postal=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}
	if value := r.FormValue("ville"); value != "" {
		updates = append(updates, "ville=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}
	if value := r.FormValue("siret"); value != "" {
		updates = append(updates, "siret=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, tiersID)
	query := "update tiers set " + strings.Join(updates, ", ") + " where tiers_id=$" + strconv.Itoa(argIndex)
	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update tier")
		return
	}

	if result.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Tier not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":  "Tier updated successfully",
		"tiers_id": tiersID,
	})
}

func supprimerTier(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	tiersID, err := strconv.Atoi(idParam)
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

	if !canManageContractDomain(ctx, conn, r, token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot delete tiers")
		return
	}

	result, err := conn.Exec(ctx, "delete from tiers where tiers_id=$1", tiersID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete tier")
		return
	}

	if result.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Tier not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":  "Tier deleted successfully",
		"tiers_id": tiersID,
	})
}

func categoriesContratsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if !canManageContractDomain(ctx, conn, r, token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot access contract categories")
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		id, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}

		var c CategorieContrat
		err = conn.QueryRow(ctx, "select categorie_contrat_id, libelle from categorie_contrat where categorie_contrat_id=$1", id).Scan(&c.CategorieContratID, &c.Libelle)
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
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "from and size are required")
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

	rows, err := conn.Query(ctx, "select categorie_contrat_id, libelle from categorie_contrat order by libelle asc offset $1 limit $2", from, size)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query categories")
		return
	}
	defer rows.Close()

	categories := make([]CategorieContrat, 0)
	for rows.Next() {
		var c CategorieContrat
		if err = rows.Scan(&c.CategorieContratID, &c.Libelle); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse categories")
			return
		}
		categories = append(categories, c)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"from":       from,
		"size":       size,
		"total":      len(categories),
		"categories": categories,
	})
}

func autocompleteCategoriesContratsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if !canManageContractDomain(ctx, conn, r, token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You cannot access contract categories")
		return
	}

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
	rows, err := conn.Query(ctx, "select categorie_contrat_id, libelle from categorie_contrat where libelle ilike $1 order by libelle asc limit $2", pattern, size)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query categories")
		return
	}
	defer rows.Close()

	categories := make([]CategorieContrat, 0)
	for rows.Next() {
		var c CategorieContrat
		if err = rows.Scan(&c.CategorieContratID, &c.Libelle); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to parse categories")
			return
		}
		categories = append(categories, c)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"query":      q,
		"size":       size,
		"total":      len(categories),
		"categories": categories,
	})
}



