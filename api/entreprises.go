package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

func isUserInEntreprise(ctx context.Context, conn *pgx.Conn, userID int, entrepriseID int) bool {
	var count int64
	err := conn.QueryRow(
		ctx,
		"select count(*) from employe where entreprise_id=$1 and client_id=$2",
		entrepriseID, userID,
	).Scan(&count)
	return err == nil && count > 0
}

func isUserManagerInEntreprise(ctx context.Context, conn *pgx.Conn, userID int, entrepriseID int) bool {
	var count int64
	err := conn.QueryRow(
		ctx,
		"select count(*) from employe where entreprise_id=$1 and manager_id=$2",
		entrepriseID, userID,
	).Scan(&count)
	return err == nil && count > 0
}

func isDirectManagerOfEmploye(ctx context.Context, conn *pgx.Conn, managerID int, entrepriseID int, employeID int) bool {
	var count int64
	err := conn.QueryRow(
		ctx,
		"select count(*) from employe where entreprise_id=$1 and client_id=$2 and manager_id=$3",
		entrepriseID, employeID, managerID,
	).Scan(&count)
	return err == nil && count > 0
}

func entreprisesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getEntreprise(w, r)
	case http.MethodPut:
		deposerEntreprise(w, r)
	case http.MethodPatch:
		modifierEntreprise(w, r)
	case http.MethodDelete:
		supprimerEntreprise(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "invalid_request", "Method not allowed")
	}
}

func employesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getEmployes(w, r)
	case http.MethodPut:
		deposerEmploye(w, r)
	case http.MethodPatch:
		modifierEmploye(w, r)
	case http.MethodDelete:
		supprimerEmploye(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "invalid_request", "Method not allowed")
	}
}

func deposerEntreprise(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	if !contains(token.Scope, "pro:entreprise_administrateur") {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only administrators can create entreprises directly. Register with account_type=entreprise to found a company as its CEO.")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	nom := r.PostForm.Get("nom")
	adresse := r.PostForm.Get("adresse")
	codePostal := r.PostForm.Get("code_postal")
	ville := r.PostForm.Get("ville")
	siret := r.PostForm.Get("siret")

	if nom == "" || adresse == "" || codePostal == "" || ville == "" || siret == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required fields: nom, adresse, code_postal, ville, siret")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	var newEntrepriseID int
	err = conn.QueryRow(
		ctx,
		"insert into entreprise (nom, adresse, code_postal, ville, siret) values ($1, $2, $3, $4, $5) returning entreprise_id",
		nom, adresse, codePostal, ville, siret,
	).Scan(&newEntrepriseID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":         "Entreprise created successfully",
		"entreprise_id":   newEntrepriseID,
		"owner_client_id": token.ClientID,
	})
}

func getEntreprise(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	query := r.URL.Query()
	idParam := query.Get("id")

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	isAdmin := contains(token.Scope, "pro:entreprise_administrateur")

	if idParam != "" {
		entrepriseID, err := strconv.Atoi(idParam)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id")
			return
		}

		if !isAdmin && !isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "You can only access your own entreprise")
			return
		}

		var entreprise Entreprise
		err = conn.QueryRow(
			ctx,
			"select entreprise_id, nom, adresse, code_postal, ville, siret from entreprise where entreprise_id=$1",
			entrepriseID,
		).Scan(&entreprise.EntrepriseID, &entreprise.Nom, &entreprise.Adresse, &entreprise.CodePostal, &entreprise.Ville, &entreprise.Siret)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "Entreprise not found"})
			return
		}

		writeJSON(w, http.StatusOK, entreprise)
		return
	}

	if !isAdmin {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Listing all entreprises requires pro:entreprise_administrateur. Query your own entreprise with ?id=")
		return
	}

	fromParam := query.Get("from")
	sizeParam := query.Get("size")

	from, err := strconv.Atoi(fromParam)
	if err != nil || from < 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid 'from' parameter")
		return
	}

	size, err := strconv.Atoi(sizeParam)
	if err != nil || size <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid 'size' parameter")
		return
	}

	rows, err := conn.Query(
		ctx,
		"select entreprise_id, nom, adresse, code_postal, ville, siret from entreprise order by entreprise_id offset $1 limit $2",
		from, size,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entreprises []Entreprise
	for rows.Next() {
		var e Entreprise
		err := rows.Scan(&e.EntrepriseID, &e.Nom, &e.Adresse, &e.CodePostal, &e.Ville, &e.Siret)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		entreprises = append(entreprises, e)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"from":        from,
		"size":        size,
		"entreprises": entreprises,
		"total":       len(entreprises),
	})
}

func modifierEntreprise(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	query := r.URL.Query()
	idParam := query.Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}

	entrepriseID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	nom := r.PostForm.Get("nom")
	adresse := r.PostForm.Get("adresse")
	codePostal := r.PostForm.Get("code_postal")
	ville := r.PostForm.Get("ville")
	siret := r.PostForm.Get("siret")

	if nom == "" && adresse == "" && codePostal == "" && ville == "" && siret == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "At least one field must be provided")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	isAdminAll := contains(token.Scope, "pro:entreprise_administrateur")
	isManagerOwn := contains(token.Scope, "pro:entreprise_manager") && isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID)

	if !isAdminAll && !isManagerOwn {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Not authorized to modify this entreprise")
		return
	}

	updates := []string{}
	args := []interface{}{entrepriseID}
	argIndex := 2

	if nom != "" {
		updates = append(updates, "nom=$"+strconv.Itoa(argIndex))
		args = append(args, nom)
		argIndex++
	}
	if adresse != "" {
		updates = append(updates, "adresse=$"+strconv.Itoa(argIndex))
		args = append(args, adresse)
		argIndex++
	}
	if codePostal != "" {
		updates = append(updates, "code_postal=$"+strconv.Itoa(argIndex))
		args = append(args, codePostal)
		argIndex++
	}
	if ville != "" {
		updates = append(updates, "ville=$"+strconv.Itoa(argIndex))
		args = append(args, ville)
		argIndex++
	}
	if siret != "" {
		updates = append(updates, "siret=$"+strconv.Itoa(argIndex))
		args = append(args, siret)
		argIndex++
	}

	updateQuery := "update entreprise set " + strings.Join(updates, ", ") + " where entreprise_id=$1"
	_, err = conn.Exec(ctx, updateQuery, args...)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Entreprise updated successfully",
	})
}

func supprimerEntreprise(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	query := r.URL.Query()
	idParam := query.Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}

	entrepriseID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	isAdminAll := contains(token.Scope, "pro:entreprise_administrateur")
	if !isAdminAll {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only administrators can delete entreprises")
		return
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "delete from employe where entreprise_id=$1", entrepriseID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	result, err := tx.Exec(ctx, "delete from entreprise where entreprise_id=$1", entrepriseID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		tx.Rollback(ctx)
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Entreprise not found"})
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Entreprise and all related employes deleted successfully",
	})
}

func deposerEmploye(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	query := r.URL.Query()
	entrepriseIDParam := query.Get("entreprise_id")
	if entrepriseIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'entreprise_id' parameter")
		return
	}

	entrepriseID, err := strconv.Atoi(entrepriseIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	clientIDParam := r.PostForm.Get("client_id")
	if clientIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'client_id' parameter")
		return
	}

	clientID, err := strconv.Atoi(clientIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid client_id")
		return
	}

	managerIDParam := r.PostForm.Get("manager_id")
	var managerID *int
	if managerIDParam != "" {
		mid, err := strconv.Atoi(managerIDParam)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid manager_id")
			return
		}
		managerID = &mid
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	isRHSupport := contains(token.Scope, "pro:rh_support")
	isRHOwn := contains(token.Scope, "pro:rh") && isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID)

	if !isRHSupport && !isRHOwn {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Not authorized to add employes to this entreprise")
		return
	}

	_, err = conn.Exec(
		ctx,
		"insert into employe (entreprise_id, client_id, manager_id) values ($1, $2, $3)",
		entrepriseID, clientID, managerID,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "Employe added successfully",
	})
}

func getEmployes(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	query := r.URL.Query()
	entrepriseIDParam := query.Get("entreprise_id")
	if entrepriseIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'entreprise_id' parameter")
		return
	}

	entrepriseID, err := strconv.Atoi(entrepriseIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id")
		return
	}

	fromParam := query.Get("from")
	sizeParam := query.Get("size")

	from, err := strconv.Atoi(fromParam)
	if err != nil || from < 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid 'from' parameter")
		return
	}

	size, err := strconv.Atoi(sizeParam)
	if err != nil || size <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid 'size' parameter")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	isAdmin := contains(token.Scope, "pro:entreprise_administrateur") || contains(token.Scope, "pro:rh_support")
	if !isAdmin && !isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You can only access employees of your own entreprise")
		return
	}

	rows, err := conn.Query(
		ctx,
		`select e.entreprise_id, e.client_id, e.manager_id, c.email, c.nom, c.prenom 
		 from employe e 
		 join client c on e.client_id = c.client_id 
		 where e.entreprise_id=$1 
		 order by e.client_id 
		 offset $2 limit $3`,
		entrepriseID, from, size,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var employes []Employe
	for rows.Next() {
		var e Employe
		err := rows.Scan(&e.EntrepriseID, &e.ClientID, &e.ManagerID, &e.Email, &e.Nom, &e.Prenom)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		employes = append(employes, e)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"from":     from,
		"size":     size,
		"employes": employes,
		"total":    len(employes),
	})
}

func modifierEmploye(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	query := r.URL.Query()
	entrepriseIDParam := query.Get("entreprise_id")
	clientIDParam := query.Get("client_id")

	if entrepriseIDParam == "" || clientIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'entreprise_id' or 'client_id' parameter")
		return
	}

	entrepriseID, err := strconv.Atoi(entrepriseIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id")
		return
	}

	clientID, err := strconv.Atoi(clientIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid client_id")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	managerIDParam := r.PostForm.Get("manager_id")
	if managerIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'manager_id' parameter")
		return
	}

	var managerID *int
	if managerIDParam != "null" {
		mid, err := strconv.Atoi(managerIDParam)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid manager_id")
			return
		}
		managerID = &mid
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	isRHSupport := contains(token.Scope, "pro:rh_support")
	isRHOwn := contains(token.Scope, "pro:rh") && isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID)
	isManagerDirect := contains(token.Scope, "pro:manager") && isDirectManagerOfEmploye(ctx, conn, token.ClientID, entrepriseID, clientID)

	if !isRHSupport && !isRHOwn && !isManagerDirect {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Not authorized to modify this employe")
		return
	}

	_, err = conn.Exec(
		ctx,
		"update employe set manager_id=$1 where entreprise_id=$2 and client_id=$3",
		managerID, entrepriseID, clientID,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Employe updated successfully",
	})
}

func supprimerEmploye(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	query := r.URL.Query()
	entrepriseIDParam := query.Get("entreprise_id")
	clientIDParam := query.Get("client_id")

	if entrepriseIDParam == "" || clientIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'entreprise_id' or 'client_id' parameter")
		return
	}

	entrepriseID, err := strconv.Atoi(entrepriseIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id")
		return
	}

	clientID, err := strconv.Atoi(clientIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid client_id")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	isRHSupport := contains(token.Scope, "pro:rh_support")
	isRHOwn := contains(token.Scope, "pro:rh") && isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID)

	if !isRHSupport && !isRHOwn {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Not authorized to remove this employe")
		return
	}

	result, err := conn.Exec(
		ctx,
		"delete from employe where entreprise_id=$1 and client_id=$2",
		entrepriseID, clientID,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "Employe not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Employe removed successfully",
	})
}

func employeAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeAPIError(w, http.StatusMethodNotAllowed, "invalid_request", "Method not allowed")
		return
	}
	rhCreateEmployeAccount(w, r)
}

func rhCreateEmployeAccount(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	entrepriseIDParam := r.PostForm.Get("entreprise_id")
	if entrepriseIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'entreprise_id' parameter")
		return
	}
	entrepriseID, err := strconv.Atoi(entrepriseIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid entreprise_id")
		return
	}

	clientEmail := r.PostForm.Get("client_email")
	clientSecret := r.PostForm.Get("client_secret")
	clientNom := r.PostForm.Get("client_nom")
	clientPrenom := r.PostForm.Get("client_prenom")
	clientTelephone := r.PostForm.Get("client_telephone")
	clientAdresse := r.PostForm.Get("client_adresse")
	clientCodePostal := r.PostForm.Get("client_code_postal")
	clientVille := r.PostForm.Get("client_ville")

	if clientEmail == "" || clientSecret == "" || clientNom == "" || clientPrenom == "" || clientTelephone == "" || clientAdresse == "" || clientCodePostal == "" || clientVille == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required fields: client_email, client_secret, client_nom, client_prenom, client_telephone, client_adresse, client_code_postal, client_ville")
		return
	}

	managerIDParam := r.PostForm.Get("manager_id")
	var managerID *int
	if managerIDParam != "" {
		mid, convErr := strconv.Atoi(managerIDParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid manager_id")
			return
		}
		managerID = &mid
	}

	requestedRoles := r.PostForm["roles"] // repeated form field: roles=pro:manager&roles=pro:rh
	grantedRoles := make([]string, 0, len(requestedRoles))
	for _, role := range requestedRoles {
		if role == CEOScope {
			writeAPIError(w, http.StatusForbidden, "access_denied", "RH cannot grant the CEO role (pro:pdg). Only account registration as account_type=entreprise creates a CEO.")
			return
		}
		if !contains(RHAssignableRoles, role) {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Unknown or non-assignable role: "+role)
			return
		}
		grantedRoles = append(grantedRoles, role)
	}

	conn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	isRHSupport := contains(token.Scope, "pro:rh_support")
	isRHOwn := contains(token.Scope, "pro:rh") && isUserInEntreprise(ctx, conn, token.ClientID, entrepriseID)
	if !isRHSupport && !isRHOwn {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Not authorized to create employe accounts for this entreprise")
		return
	}

	var count int64
	err = conn.QueryRow(ctx, "select count(client_id) from client where email=$1", clientEmail).Scan(&count)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Email already exists.")
		return
	}

	hashedPassword := hashPassword(clientSecret)
	if hashedPassword == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var newClientID int
	err = tx.QueryRow(
		ctx,
		"insert into client (nom, prenom, email, password, telephone, adresse, code_postal, ville, score) values ($1, $2, $3, $4, $5, $6, $7, $8, 0) returning client_id",
		clientNom, clientPrenom, clientEmail, hashedPassword, clientTelephone, clientAdresse, clientCodePostal, clientVille,
	).Scan(&newClientID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(
		ctx,
		"insert into employe (entreprise_id, client_id, manager_id) values ($1, $2, $3)",
		entrepriseID, newClientID, managerID,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(grantedRoles) > 0 {
		if err := grantRoles(ctx, tx, newClientID, grantedRoles); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":       "Employe account created successfully",
		"client_id":     newClientID,
		"entreprise_id": entrepriseID,
		"roles":         grantedRoles,
	})
}
