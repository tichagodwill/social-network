package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"social-network/api"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

// middleware that will check the existance of the cookie on each handler
func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the cookie from the browser
		cookie, err := r.Cookie("AccessToken")
		if err != nil {
			// check if the cookie exists from the browser
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthenticated user", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Something went wrong", http.StatusUnauthorized)
			return
		}

		// get the value of the cookie
		cookieValue := cookie.Value

		// check if the cookie exists in the already active sessions
		if _, ok := util.UserSession[cookieValue]; !ok {
			http.Error(w, "Unauthorized user", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Open the database connection
	err := sqlite.OpenDB("./social-network.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer sqlite.DB.Close()

	var arg string

	// check if an argument is passed
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	// check case insesitive
	if strings.EqualFold(arg, "flush") {
		// remove all data from the database
		if err := sqlite.ClearDatabase(); err != nil {
			log.Fatalf("Error flushing database: %v", err)
		}
	} else if strings.EqualFold(arg, "rollback") {
		// roll back the migrations
		if err := sqlite.RollbackMigrations(); err != nil {
			log.Fatalf("Error rolling back: %v", err)
		}

		// exit after rolling back
		return
	}

	mux := http.NewServeMux()

	//NOTE: GO VERSION 1.22+ WILL BE USED IN THIS PROJECT IF YOU DON'T HAVE THAT PLEASE UPDATE YOUR GO
	mux.HandleFunc("POST /register", api.RegisterHandler)
	mux.HandleFunc("POST /login", api.LoginHandler)
	mux.HandleFunc("POST /logout", api.LogoutHandler)

	mux.Handle("POST /posts", middleware(http.HandlerFunc(api.CreatePost)))
	mux.Handle("GET /posts/{id}", middleware(http.HandlerFunc(api.ViewPost)))
	mux.Handle("GET /posts", middleware(http.HandlerFunc(api.GetPosts)))

	mux.Handle("POST /comments", middleware(http.HandlerFunc(api.CreateComment)))
	mux.Handle("GET /comments/{postID}", middleware(http.HandlerFunc(api.GetComments)))

	mux.Handle("GET /groups", middleware(http.HandlerFunc(api.VeiwGorups)))
	mux.Handle("POST /groups", middleware(http.HandlerFunc(api.CreateGroup)))
	mux.Handle("POST /groups/{id}/posts", middleware(http.HandlerFunc(api.CreateGroupPost)))
	mux.Handle("GET /groups/{id}/posts", middleware(http.HandlerFunc(api.GetGroupPost)))

	fmt.Println("Server running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
