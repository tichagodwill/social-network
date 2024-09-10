package main

import (
	"fmt"
	"log"
	"net/http"
	"social-network/api"
	"social-network/pkg/db/sqlite"
)

func main() {
	// Open the database connection
	err := sqlite.OpenDB("./social-network.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer sqlite.DB.Close()

	mux := http.NewServeMux()

	//NOTE: GO VERSION 1.22+ WILL BE USED IN THIS PROJECT IF YOU DON'T HAVE THAT PLEASE UPDATE YOUR GO
	mux.HandleFunc("POST /register", api.RegisterHandler)
	mux.HandleFunc("POST /login", api.LoginHandler)
	mux.HandleFunc("POST /logout", api.LogoutHandler)

	fmt.Println("Server running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
