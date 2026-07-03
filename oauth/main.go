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
	http.HandleFunc("/oauth/v3/inscription", createAccount)
	http.HandleFunc("/oauth/v3/token", token)
	http.HandleFunc("/oauth/v3/introspect", introspect)
	http.HandleFunc("/oauth/v3/roles", roles)

	port := getEnv("PORT", "8080")
	fmt.Println("Listening on :" + port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
