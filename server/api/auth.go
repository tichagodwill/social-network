package api

import (
	"fmt"
	"net/http"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("user regsiter")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("user login")
}
