package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)


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

	if strings.HasPrefix(authHeader, "Basic ") {
		writeAPIError(w, http.StatusBadRequest, "duplicate_credentials", "You haven't provided a token to introspect.")
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

	writeJSON(w, http.StatusOK, map[string]any{
		"active":     true,
		"scope":      "",
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

	conn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

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

	var newClientID int
	err = conn.QueryRow(
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

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":   "Account created successfully",
		"client_id": newClientID,
	})
}
