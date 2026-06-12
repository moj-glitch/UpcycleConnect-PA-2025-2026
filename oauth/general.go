package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const OAUTH_URL = "http://localhost/oauth/v3/introspect"
const DATA_DIR = "C:/Users/tomto/Documents/go/oauth/DATA"

type Thread struct {
	ThreadID   int    `json:"thread_id"`
	Categorie  int    `json:"categorie_thread"`
	Titre      string `json:"titre"`
	Message    string `json:"message"`
	Resolu     string `json:"resolu"`
	DateCreate string `json:"date_creation"`
}

type MessageThread struct {
	MessageThreadID int    `json:"message_thread_id"`
	ThreadID        int    `json:"thread_id"`
	ClientID        int    `json:"client_id"`
	Message         string `json:"message"`
	Parent          *int   `json:"parent"`
	DateEnvoi       string `json:"date_envoi"`
}

type Annonce struct {
	AnnonceID        int     `json:"annonce_id"`
	Vendeur          string  `json:"vendeur"`
	Acheteur         string  `json:"acheteur"`
	Categorie        string  `json:"categorie"`
	Titre            string  `json:"titre"`
	Prix             float64 `json:"prix"`
	Description      string  `json:"description"`
	Etat             string  `json:"etat"`
	Taxe             float64 `json:"taxe"`
	Image            string  `json:"image"`
	Barcode          string  `json:"barcode"`
	Date_Publication string  `json:"date_publication"`
}

type IntrospectionPayload struct {
	Active   bool     `json:"active"`
	Scope    []string `json:"scope"`
	ClientID int      `json:"client_id"`
	Username string   `json:"username"`
}

type CategorieAnnonce struct {
	CategorieID int    `json:"categorie_id"`
	Libelle     string `json:"libelle"`
}

type Entreprise struct {
	EntrepriseID int    `json:"entreprise_id"`
	Nom          string `json:"nom"`
	Adresse      string `json:"adresse"`
	CodePostal   string `json:"code_postal"`
	Ville        string `json:"ville"`
	Siret        string `json:"siret"`
}

type Employe struct {
	EntrepriseID int    `json:"entreprise_id"`
	ClientID     int    `json:"client_id"`
	ManagerID    *int   `json:"manager_id"`
	Email        string `json:"email"`
	Nom          string `json:"nom"`
	Prenom       string `json:"prenom"`
}

type Evenement struct {
	EvenementID int    `json:"evenement_id"`
	Date        string `json:"date"`
	Nom         string `json:"nom"`
	Description string `json:"description"`
	Statut      string `json:"statut"`
	Categorie   int    `json:"categorie"`
}

type Tutoriel struct {
	TutorielID   int    `json:"tutoriel_id"`
	Titre        string `json:"titre"`
	DateCreation string `json:"date_creation"`
	Article      string `json:"article"`
	Video        string `json:"video"`
	Categorie    int    `json:"categorie"`
}

type Contrat struct {
	ContratID          int     `json:"contrat_id"`
	EntrepriseID       int     `json:"entreprise_id"`
	CategorieContratID *int    `json:"categorie_contrat_id"`
	Description        string  `json:"description"`
	Depense            float64 `json:"depense"`
	Gain               float64 `json:"gain"`
	PDF                string  `json:"pdf"`
	DateDebut          string  `json:"date_debut"`
	DateFin            *string `json:"date_fin"`
}

type Tier struct {
	TiersID    int     `json:"tiers_id"`
	Nom        string  `json:"nom"`
	Adresse    string  `json:"adresse"`
	CodePostal string  `json:"code_postal"`
	Ville      string  `json:"ville"`
	Siret      *string `json:"siret"`
}

type CategorieContrat struct {
	CategorieContratID int    `json:"categorie_contrat_id"`
	Libelle            string `json:"libelle"`
}

type Projet struct {
	ProjetID        int    `json:"projet_id"`
	EntrepriseID    int    `json:"entreprise_id"`
	CategorieProjet int    `json:"categorie_projet"`
	Titre           string `json:"titre"`
	Image           string `json:"image"`
	Video           string `json:"video"`
	Texte           string `json:"texte"`
	DateCreation    string `json:"date_creation"`
}

type CategorieProjet struct {
	CategorieProjetID int    `json:"categorie_projet_id"`
	Libelle           string `json:"libelle"`
}

type Materiau struct {
	MaterieauID            int     `json:"materieau_id"`
	Nom                    string  `json:"nom"`
	Densite                float64 `json:"densite"`
	PrixKG                 float64 `json:"prix_kg"`
	Durete                 float64 `json:"durete"`
	FonteDegree            float64 `json:"fonte_degree"`
	ElasticiteYoung        float64 `json:"elasticite_young"`
	ResistanceTorsion      float64 `json:"resistance_torsion"`
	ResistanceCompression  float64 `json:"resistance_compression"`
	Resilience             float64 `json:"resilience"`
	ConductiviteThermique  float64 `json:"conductivite_thermique"`
	ConductiviteElectrique float64 `json:"conductivite_electrique"`
	DilatationThermique    float64 `json:"dilatation_thermique"`
	Porosite               float64 `json:"porosite"`
	Opacite                float64 `json:"opacite"`
	Magnetisme             float64 `json:"magnetisme"`
	Accoustique            float64 `json:"accoustique"`
	Recyclable             string  `json:"recyclable"`
	EmpreinteCO2           float64 `json:"empreinte_co2"`
	Toxicite               float64 `json:"toxicite"`
}

type ClientRolesResponse struct {
	ClientID				string `json:"client_id"`
	Roles					[]string `json:"roles"`
}

func tryAuth(w http.ResponseWriter, r *http.Request) *IntrospectionPayload {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if contentType != "" && !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") && !strings.HasPrefix(contentType, "multipart/form-data") {
		writeAPIError(w, http.StatusBadRequest, "invalid_client", "The URI does not support the requested content type.")
		return &IntrospectionPayload{false, nil, 0, ""}
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		writeAPIError(w, http.StatusBadRequest, "invalid_client", "The client failed to authenticate with the authorization server.")
		return &IntrospectionPayload{false, nil, 0, ""}
	}

	if strings.HasPrefix(authHeader, "Basic ") {
		writeAPIError(w, http.StatusBadRequest, "duplicate_credentials", "You haven't provided a token to introspect.")
		return &IntrospectionPayload{false, nil, 0, ""}
	}

	req, err := http.NewRequest("GET", OAUTH_URL, nil)
	req.Header.Add("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "verify your access token")
		return &IntrospectionPayload{false, nil, 0, ""}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "verify your access token")
		return &IntrospectionPayload{false, nil, 0, ""}
	}

	raw := make(map[string]any)
	err = json.Unmarshal(body, &raw)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "verify your access token")
		return &IntrospectionPayload{false, nil, 0, ""}
	}

	jsonData := IntrospectionPayload{Active: false, Scope: []string{}, ClientID: 0, Username: ""}

	if active, ok := raw["active"].(bool); ok {
		jsonData.Active = active
	}

	if username, ok := raw["username"].(string); ok {
		jsonData.Username = username
	}

	switch scope := raw["scope"].(type) {
	case []any:
		for _, item := range scope {
			if s, ok := item.(string); ok {
				jsonData.Scope = append(jsonData.Scope, s)
			}
		}
	case []string:
		jsonData.Scope = scope
	case string:
		if scope != "" {
			jsonData.Scope = strings.Fields(scope)
		}
	}

	switch clientID := raw["client_id"].(type) {
	case float64:
		jsonData.ClientID = int(clientID)
	case int:
		jsonData.ClientID = clientID
	case string:
		parsedID, convErr := strconv.Atoi(clientID)
		if convErr == nil {
			jsonData.ClientID = parsedID
		}
	}

	if !jsonData.Active || jsonData.ClientID == 0 {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "verify your access token")
		return &IntrospectionPayload{false, nil, 0, ""}
	}

	return &jsonData
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func generateTokenValue() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAPIError(w http.ResponseWriter, status int, errCode string, description string) {
	writeJSON(w, status, map[string]any{
		"error":             errCode,
		"error_description": description,
	})
}
