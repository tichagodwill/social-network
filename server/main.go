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
	"social-network/middleware"
)

// authMiddleware checks the existence of the cookie on each handler
func authMiddleware(next http.Handler) http.Handler {
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
		return
	} else if strings.EqualFold(arg, "migrate") {
		// run migrations
		if err := sqlite.RunMigrations(); err != nil {
			log.Fatalf("Error running migrations: %v", err)
		}
		return
	} else if strings.EqualFold(arg, "show-data") {
		sqlite.PrintDatabaseContent()
		return
	}

	mux := http.NewServeMux()

	// Public routes (no middleware)
	mux.HandleFunc("POST /register", api.RegisterHandler)
	mux.HandleFunc("POST /login", api.LoginHandler)
	mux.HandleFunc("POST /logout", api.LogoutHandler)
	mux.HandleFunc("GET /user/current", api.GetCurrentUser)

	// Protected routes (with authMiddleware)
	mux.Handle("POST /posts", authMiddleware(http.HandlerFunc(api.CreatePost)))
	mux.Handle("GET /posts/{id}", authMiddleware(http.HandlerFunc(api.ViewPost)))
	mux.Handle("GET /posts", authMiddleware(http.HandlerFunc(api.GetPosts)))

	mux.Handle("POST /comments", authMiddleware(http.HandlerFunc(api.CreateComment)))
	mux.Handle("GET /comments/{postID}", authMiddleware(http.HandlerFunc(api.GetComments)))

	mux.Handle("GET /groups", authMiddleware(http.HandlerFunc(api.VeiwGorups)))
	mux.Handle("POST /groups", authMiddleware(http.HandlerFunc(api.CreateGroup)))
	mux.Handle("POST /groups/{id}/posts", authMiddleware(http.HandlerFunc(api.CreateGroupPost)))
	mux.Handle("GET /groups/{id}/posts", authMiddleware(http.HandlerFunc(api.GetGroupPost)))
	mux.Handle("POST /groups/invitation", authMiddleware(http.HandlerFunc(api.GroupInvitation)))
	mux.Handle("POST /groups/accept", authMiddleware(http.HandlerFunc(api.GroupAccept)))
	mux.Handle("POST /groups/reject", authMiddleware(http.HandlerFunc(api.GroupReject)))
	mux.Handle("POST /groups/leave", authMiddleware(http.HandlerFunc(api.GroupLeave)))

	mux.Handle("POST /follow", authMiddleware(http.HandlerFunc(api.RequestFollowUser)))
	mux.Handle("PATCH /follow/{requestID}", authMiddleware(http.HandlerFunc(api.AcceptOrRejectRequest)))
	mux.Handle("GET /follower/{userID}", authMiddleware(http.HandlerFunc(api.GetFollowers)))

	mux.Handle("GET /user/{userID}", authMiddleware(http.HandlerFunc(api.UserProfile)))

	mux.Handle("/ws", authMiddleware(http.HandlerFunc(api.WebSocketHandler)))

	mux.Handle("GET /groups/{id}", authMiddleware(http.HandlerFunc(api.GetGroup)))

	mux.Handle("GET /notifications", authMiddleware(http.HandlerFunc(api.GetNotifications)))

	mux.Handle("GET /groups/{id}/members", authMiddleware(http.HandlerFunc(api.GetGroupMembers)))

	mux.Handle("GET /groups/{id}/events", authMiddleware(http.HandlerFunc(api.GetGroupEvents)))
	mux.Handle("POST /groups/{id}/events", authMiddleware(http.HandlerFunc(api.CreateGroupEvent)))
	mux.Handle("POST /groups/events/{eventId}/respond", authMiddleware(http.HandlerFunc(api.RespondToGroupEvent)))

	mux.Handle("PUT /groups/{id}", authMiddleware(http.HandlerFunc(api.UpdateGroup)))
	mux.Handle("DELETE /groups/{id}", authMiddleware(http.HandlerFunc(api.DeleteGroup)))

	mux.Handle("PUT /groups/{id}/members/{memberId}/role", authMiddleware(http.HandlerFunc(api.UpdateMemberRole)))
	mux.Handle("DELETE /groups/{id}/members/{memberId}", authMiddleware(http.HandlerFunc(api.RemoveMember)))

	mux.Handle("GET /groups/{id}/requests", authMiddleware(http.HandlerFunc(api.GetGroupRequests)))

	// Wrap the entire mux with CORS middleware
	handler := middleware.CORS(mux)

	fmt.Println("Server running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
