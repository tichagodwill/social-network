package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	//NOTE: GO VERSION 1.22+ WILL BE USED IN THIS PROJECT IF YOU DON'T HAVE THAT PLEASE UDPATE YOUR GO
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Home")
	})

	fmt.Println("Server running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
