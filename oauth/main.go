package main

import (
	"fmt"
	"net/http"
)

const DATABASE_AUTH_URL string = "postgres://postgres:035464@localhost:5432/database"
const DATABASE_DATA_URL string = "postgres://postgres:035464@localhost:5432/database"
const DATABASE_PUBLIC_URL string = "postgres://postgres:035464@localhost:5432/database"
const DATABASE_PRO_URL string = "postgres://postgres:035464@localhost:5432/database"
const DATABASE_ADMIN_URL string = "postgres://postgres:035464@localhost:5432/database"

func main() {
	http.HandleFunc("/oauth/v3/inscription", createAccount)
	http.HandleFunc("/oauth/v3/token", token)
	http.HandleFunc("/oauth/v3/introspect", introspect)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
