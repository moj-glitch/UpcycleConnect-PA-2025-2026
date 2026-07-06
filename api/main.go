package main

import (
	"fmt"
	"net/http"
)

var (
	DATABASE_AUTH_URL   = getEnv("DATABASE_AUTH_URL", "postgres://postgres:035464@localhost:5432/database")
	DATABASE_DATA_URL   = getEnv("DATABASE_DATA_URL", DATABASE_AUTH_URL)
	DATABASE_PUBLIC_URL = getEnv("DATABASE_PUBLIC_URL", DATABASE_AUTH_URL)
	DATABASE_PRO_URL    = getEnv("DATABASE_PRO_URL", DATABASE_AUTH_URL)
	DATABASE_ADMIN_URL  = getEnv("DATABASE_ADMIN_URL", DATABASE_AUTH_URL)
)

func main() {
	http.HandleFunc("/api/v1/evenements", evenementsHandler)
	http.HandleFunc("/api/v1/tutoriels", tutorielsHandler)
	http.HandleFunc("/api/v1/annonces", annoncesHandler)
	http.HandleFunc("/api/v1/annonces/messages", messagesAnnonceHandler)
	http.HandleFunc("/api/v1/annonces/categories", categoriesAnnoncesHandler)
	http.HandleFunc("/api/v1/threads", threadsHandler)
	http.HandleFunc("/api/v1/threads/messages", messagesThreadHandler)
	http.HandleFunc("/api/v1/threads/messages/children", messagesThreadChildrenHandler)
	http.HandleFunc("/api/v1/entreprises", entreprisesHandler)
	http.HandleFunc("/api/v1/entreprises/employes", employesHandler)
	http.HandleFunc("/api/v1/entreprises/employes/inscription", employeAccountHandler)
	http.HandleFunc("/api/v1/contrats", contratsHandler)
	http.HandleFunc("/api/v1/tiers", tiersHandler)
	http.HandleFunc("/api/v1/tiers/autocomplete", autocompleteTiersHandler)
	http.HandleFunc("/api/v1/contrats/categories", categoriesContratsHandler)
	http.HandleFunc("/api/v1/contrats/categories/autocomplete", autocompleteCategoriesContratsHandler)
	http.HandleFunc("/api/v1/projets", projetsHandler)
	http.HandleFunc("/api/v1/projets/categories", categoriesProjetsHandler)
	http.HandleFunc("/api/v1/projets/categories/autocomplete", autocompleteCategoriesProjetsHandler)
	http.HandleFunc("/api/v1/materiaux", materiauxHandler)

	port := getEnv("PORT", "8081")
	fmt.Println("Listening on :" + port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
