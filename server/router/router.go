package router

import (
	"net/http"
	"social-network/api"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	// Auth routes
	router.HandleFunc("/register", api.Register).Methods("POST")
	router.HandleFunc("/login", api.Login).Methods("POST")
	router.HandleFunc("/logout", api.Logout).Methods("POST")
	router.HandleFunc("/session", api.CheckSession).Methods("GET")

	// Group routes
	router.HandleFunc("/groups", api.ViewGroups).Methods("GET")
	router.HandleFunc("/groups", api.CreateGroup).Methods("POST")
	router.HandleFunc("/groups/{id}", api.GetGroup).Methods("GET")
	router.HandleFunc("/groups/{id}", api.UpdateGroup).Methods("PUT")
	router.HandleFunc("/groups/{id}", api.DeleteGroup).Methods("DELETE")
	
	// Group membership routes
	router.HandleFunc("/groups/{id}/members", api.GetGroupMembers).Methods("GET")
	router.HandleFunc("/groups/{id}/join", api.RequestJoinGroup).Methods("POST")
	router.HandleFunc("/groups/invitation", api.InviteMember).Methods("POST", "OPTIONS")
	router.HandleFunc("/groups/invitation/{id}/{action}", api.HandleInvitation).Methods("POST")
	router.HandleFunc("/groups/{id}/invitation/status", api.GetInvitationStatus).Methods("GET")
	router.HandleFunc("/groups/{id}/requests", api.GetGroupRequests).Methods("GET")
	router.HandleFunc("/groups/{id}/members/{memberId}", api.RemoveMember).Methods("DELETE")
	router.HandleFunc("/groups/{id}/members/{memberId}/role", api.UpdateMemberRole).Methods("PUT")

	// Group content routes
	router.HandleFunc("/groups/{id}/posts", api.GetGroupPosts).Methods("GET")
	router.HandleFunc("/groups/{id}/posts", api.CreateGroupPost).Methods("POST")
	router.HandleFunc("/groups/{id}/posts/{postId}/comments", api.CreateGroupPostComment).Methods("POST")
	router.HandleFunc("/groups/{id}/events", api.GetGroupEvents).Methods("GET")
	router.HandleFunc("/groups/{id}/events", api.CreateGroupEvent).Methods("POST")
	router.HandleFunc("/groups/events/{eventId}/respond", api.RespondToGroupEvent).Methods("POST")

	// Add CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	return router
}

func Serve(router *mux.Router) {
	http.ListenAndServe(":8080", router)
} 