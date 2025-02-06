package routes

import (
	"net/http"
	"social-network/api"
)

func SetupRoutes(mux *http.ServeMux) {
	// Group routes
	mux.HandleFunc("GET /groups", api.ViewGroups)
	mux.HandleFunc("POST /groups", api.CreateGroup)
	mux.HandleFunc("GET /groups/{id}", api.GetGroup)
	
	// Group invitation routes
	mux.HandleFunc("POST /groups/{id}/invitations", api.InviteToGroup)
	mux.HandleFunc("POST /groups/{id}/invitations/{invitationId}/{action}", api.HandleInvitation)
	
// Group membership routes
mux.HandleFunc("GET /groups/{id}/members", api.GetGroupMembers)

// Chat routes
mux.HandleFunc("POST /chat/direct", api.CreateOrGetDirectChat)

// Notification routes
	mux.HandleFunc("GET /notifications", api.GetNotifications)
	mux.HandleFunc("GET /notifications/{id}/read", api.MarkNotificationAsRead)
}
