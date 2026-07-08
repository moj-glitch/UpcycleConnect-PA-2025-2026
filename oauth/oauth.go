package main

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// readBasicCredentials decodes a "Basic <base64(identifier:secret)>" header
// into its two parts. This replaces the two previously-duplicated functions
// (readBasicClientCredentials / readBasicEmailCredentials) that did exactly
// the same thing.
func readBasicCredentials(authHeader string) (string, string, error) {
	if authHeader == "" {
		return "", "", http.ErrNoCookie
	}

	authParts := strings.SplitN(authHeader, " ", 2)
	if len(authParts) != 2 || authParts[0] != "Basic" {
		return "", "", http.ErrNoCookie
	}

	basicDecoded, err := base64.StdEncoding.DecodeString(authParts[1])
	if err != nil {
		return "", "", err
	}

	creds := strings.SplitN(string(basicDecoded), ":", 2)
	if len(creds) != 2 {
		return "", "", http.ErrNoCookie
	}

	return creds[0], creds[1], nil
}

func isEmailIdentifier(identifier string) bool {
	if identifier == "" {
		return false
	}

	_, err := mail.ParseAddress(identifier)
	return err == nil
}

// token issues an access token. The client authenticates with HTTP Basic
// auth where the username is EITHER their email OR their numeric client_id -
// isEmailIdentifier decides which lookup to run.
func token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	clientIdentifier, secret, err := readBasicCredentials(r.Header.Get("Authorization"))
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	var hashedSecret string
	var clientID int

	conn, err := pgx.Connect(context.Background(), DATABASE_AUTH_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())

	if isEmailIdentifier(clientIdentifier) {
		err = conn.QueryRow(context.Background(), "select client_id, password from client where email=$1", clientIdentifier).Scan(&clientID, &hashedSecret)
	} else {
		clientID, err = strconv.Atoi(clientIdentifier)
		if err == nil {
			err = conn.QueryRow(context.Background(), "select password from client where client_id=$1", clientID).Scan(&hashedSecret)
		}
	}
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(secret)); err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	accessToken, err := generateTokenValue()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = conn.Exec(
		context.Background(),
		"insert into token (client_id, token_value, date_expiration, active, revoked) values ($1, $2, now() + interval '1 hour', '1', '0')",
		clientID,
		accessToken,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        "",
	})
}

func roles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	email, secret, err := readBasicCredentials(r.Header.Get("Authorization"))
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	var hashedSecret string
	var clientID int
	err = conn.QueryRow(ctx, "select client_id, password from client where email=$1", email).Scan(&clientID, &hashedSecret)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(secret)); err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	rows, err := conn.Query(ctx, "select role.libelle from possede join role on possede.role_id = role.role_id where possede.client_id = $1 order by role.role_id asc", clientID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	roleList := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		roleList = append(roleList, role)
	}

	writeJSON(w, http.StatusOK, ClientRolesResponse{
		ClientID: strconv.Itoa(clientID),
		Roles:    roleList,
	})
}

func introspect(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		writeAPIError(w, http.StatusBadRequest, "invalid_client", "The URI does not support the requested content type.")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		writeAPIError(w, http.StatusBadRequest, "invalid_client", "The client failed to authenticate with the authorization server.")
		return
	}

	tokenToCheck := strings.TrimPrefix(authHeader, "Bearer ")

	conn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	var tokenID int64
	var tokenExpiration string
	var tokenActive string
	var tokenRevoked string

	err = conn.QueryRow(ctx, "select token_id, date_expiration, active, revoked from token where token_value=$1", tokenToCheck).Scan(&tokenID, &tokenExpiration, &tokenActive, &tokenRevoked)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"active":    false,
			"scope":     nil,
			"client_id": nil,
			"username":  nil,
			"exp":       nil,
			"revoked":   nil,
			"iat":       nil,
			"sub":       nil,
			"aud":       nil,
			"iss":       nil,
		})
		return
	}

	if tokenActive == "0" || tokenRevoked == "1" {
		writeJSON(w, http.StatusOK, map[string]any{
			"active":    false,
			"scope":     nil,
			"client_id": nil,
			"username":  nil,
			"exp":       tokenExpiration,
			"revoked":   tokenRevoked,
			"iat":       nil,
			"sub":       nil,
			"aud":       nil,
			"iss":       nil,
		})
		return
	}

	var clientID int
	var email string

	err = conn.QueryRow(ctx, "select token.client_id, client.email from token join client on token.client_id = client.client_id where token.token_id=$1", tokenID).Scan(&clientID, &email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	scopeRows, err := conn.Query(ctx, "select role.libelle from possede join role on possede.role_id = role.role_id where possede.client_id = $1", clientID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer scopeRows.Close()

	var scopes []string
	for scopeRows.Next() {
		var libelle string
		if err := scopeRows.Scan(&libelle); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		scopes = append(scopes, libelle)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"active":     true,
		"scope":      strings.Join(scopes, " "),
		"client_id":  clientID,
		"username":   email,
		"token_type": "Bearer",
		"exp":        tokenExpiration,
		"iat":        nil,
		"sub":        nil,
		"aud":        nil,
		"iss":        nil,
	})
}

// createAccount registers a new client. The caller picks account_type:
//   - "freemium" (default if omitted): a plain client account, no company.
//   - "entreprise": also creates a new entreprise, attaches the new client
//     to it as an employee with no manager, and grants the full set of
//     company-scoped roles (CompanyScopedRoles, including CEOScope) so the
//     registrant is immediately the CEO of the company they just founded.
//
// Everything below happens in a single transaction: if any step fails,
// nothing is created - you never end up with a client that has no
// entreprise, or an entreprise that has no CEO.
func createAccount(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body.")
		return
	}

	clientEmail := r.PostForm.Get("client_email")
	clientSecret := r.PostForm.Get("client_secret")
	clientConfirm := r.PostForm.Get("client_confirm")
	clientNom := r.PostForm.Get("client_nom")
	clientPrenom := r.PostForm.Get("client_prenom")
	clientTelephone := r.PostForm.Get("client_telephone")
	clientAdresse := r.PostForm.Get("client_adresse")
	clientCodePostal := r.PostForm.Get("client_code_postal")
	clientVille := r.PostForm.Get("client_ville")
	clientSiret := r.PostForm.Get("client_siret")

	if clientEmail == "" || clientSecret == "" || clientNom == "" || clientPrenom == "" || clientTelephone == "" || clientAdresse == "" || clientCodePostal == "" || clientVille == "" || clientSiret == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required fields: client_email, client_secret, client_nom, client_prenom, client_telephone, client_adresse, client_code_postal, client_ville, client_siret")
		return
	}

	if clientSecret != clientConfirm {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Passwords do not match.")
		return
	}

	if !telephoneRegexp.MatchString(clientTelephone) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "client_telephone must be in international format, e.g. +33612345678")
		return
	}
	if !codePostalRegexp.MatchString(clientCodePostal) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "client_code_postal must be exactly 5 digits")
		return
	}
	if !siretRegexp.MatchString(clientSiret) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "client_siret must be exactly 14 digits")
		return
	}

	accountType := strings.ToLower(strings.TrimSpace(r.PostForm.Get("account_type")))
	if accountType == "" {
		accountType = "freemium"
	}
	if accountType != "freemium" && accountType != "entreprise" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "account_type must be 'freemium' or 'entreprise'")
		return
	}

	forfaitLibelle := "Freemium"
	if accountType == "entreprise" {
		forfaitChoice := strings.ToLower(strings.TrimSpace(r.PostForm.Get("forfait")))
		if forfaitChoice == "" {
			forfaitChoice = "gratuit"
		}
		switch forfaitChoice {
		case "gratuit":
			forfaitLibelle = "Pro Gratuit"
		case "payant":
			forfaitLibelle = "Pro Payant"
		default:
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "forfait must be 'gratuit' or 'payant'")
			return
		}
	}

	var entrepriseNom, entrepriseAdresse, entrepriseCodePostal, entrepriseVille, entrepriseSiret string
	if accountType == "entreprise" {
		entrepriseNom = r.PostForm.Get("entreprise_nom")
		entrepriseAdresse = r.PostForm.Get("entreprise_adresse")
		entrepriseCodePostal = r.PostForm.Get("entreprise_code_postal")
		entrepriseVille = r.PostForm.Get("entreprise_ville")
		entrepriseSiret = r.PostForm.Get("entreprise_siret")

		if entrepriseNom == "" || entrepriseAdresse == "" || entrepriseCodePostal == "" || entrepriseVille == "" || entrepriseSiret == "" {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "account_type=entreprise requires entreprise_nom, entreprise_adresse, entreprise_code_postal, entreprise_ville, entreprise_siret")
			return
		}
		if !codePostalRegexp.MatchString(entrepriseCodePostal) {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "entreprise_code_postal must be exactly 5 digits")
			return
		}
		if !siretRegexp.MatchString(entrepriseSiret) {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "entreprise_siret must be exactly 14 digits")
			return
		}
	}

	conn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	var count int64
	err = conn.QueryRow(ctx, "select count(client_id) from client where email=$1 or siret=$2", clientEmail, clientSiret).Scan(&count)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "A client with this email or siret already exists.")
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
		"insert into client (nom, prenom, email, password, telephone, adresse, code_postal, ville, score, siret) values ($1, $2, $3, $4, $5, $6, $7, $8, 0, $9) returning client_id",
		clientNom,
		clientPrenom,
		clientEmail,
		hashedPassword,
		clientTelephone,
		clientAdresse,
		clientCodePostal,
		clientVille,
		clientSiret,
	).Scan(&newClientID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"message":      "Account created successfully",
		"client_id":    newClientID,
		"account_type": accountType,
	}

	if accountType == "entreprise" {
		var entrepriseCount int64
		err = tx.QueryRow(ctx, "select count(entreprise_id) from entreprise where nom=$1 or siret=$2", entrepriseNom, entrepriseSiret).Scan(&entrepriseCount)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if entrepriseCount > 0 {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "An entreprise with this nom or siret already exists.")
			return
		}

		var newEntrepriseID int
		err = tx.QueryRow(
			ctx,
			"insert into entreprise (nom, adresse, code_postal, ville, siret) values ($1, $2, $3, $4, $5) returning entreprise_id",
			entrepriseNom, entrepriseAdresse, entrepriseCodePostal, entrepriseVille, entrepriseSiret,
		).Scan(&newEntrepriseID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(
			ctx,
			"insert into employe (entreprise_id, client_id, manager_id) values ($1, $2, null)",
			newEntrepriseID, newClientID,
		)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := grantRoles(ctx, tx, newClientID, CompanyScopedRoles); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		response["entreprise_id"] = newEntrepriseID
		response["roles"] = CompanyScopedRoles
	}

	_, err = tx.Exec(
		ctx,
		"insert into souscrit (client_id, forfait_id) select $1, forfait_id from forfait where libelle=$2",
		newClientID, forfaitLibelle,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if forfaitLibelle == "Pro Gratuit" {
		_, err = tx.Exec(
			ctx,
			"insert into souscrit (client_id, forfait_id) select $1, forfait_id from forfait where libelle='Freemium'",
			newClientID,
		)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	response["forfait"] = forfaitLibelle

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}
