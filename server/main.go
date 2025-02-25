package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"social-network/api"
	"social-network/middleware"
	"social-network/pkg/db/sqlite"
	"social-network/util"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Add file and line number to logs
	log.Println("Server starting...")

	// Open the database connection
	err := sqlite.OpenDB("./social-network.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqlite.DB.Close()

	// Run migrations
	if err := sqlite.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

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
	mux.Handle("GET /posts/{id}/details", authMiddleware(http.HandlerFunc(api.GetPostDetails)))
	mux.Handle("POST /posts/addComment", authMiddleware(http.HandlerFunc(api.AddPostComment)))

	mux.Handle("POST /comments", authMiddleware(http.HandlerFunc(api.CreateComment)))
	mux.Handle("GET /comments/{postID}", authMiddleware(http.HandlerFunc(api.GetComments)))

	//explore page
	mux.Handle("POST /explore", authMiddleware(http.HandlerFunc(api.GetExplore)))

	// Group these related routes together and order them from most specific to the least specific
	// Basic group routes
	mux.Handle("GET /groups", authMiddleware(http.HandlerFunc(api.ViewGroups)))
	mux.Handle("POST /groups", authMiddleware(http.HandlerFunc(api.CreateGroup)))
	mux.Handle("GET /groups/{id}", authMiddleware(http.HandlerFunc(api.GetGroup)))
	mux.Handle("PUT /groups/{id}", authMiddleware(http.HandlerFunc(api.UpdateGroup)))
	mux.Handle("DELETE /groups/{id}", authMiddleware(http.HandlerFunc(api.DeleteGroup)))

	// Group member management
	mux.Handle("GET /groups/{id}/members", authMiddleware(http.HandlerFunc(api.GetGroupMembers)))
	mux.Handle("GET /groups/{id}/members/role", authMiddleware(http.HandlerFunc(api.GetMemberRole)))
	mux.Handle("PUT /groups/{id}/members/{memberId}/role", authMiddleware(http.HandlerFunc(api.UpdateMemberRole)))
	mux.Handle("DELETE /groups/{id}/members/{memberId}", authMiddleware(http.HandlerFunc(api.RemoveMember)))

	// Group invitation routes
	mux.Handle("POST /groups/{id}/invitations", authMiddleware(http.HandlerFunc(api.InviteToGroup)))
	mux.Handle("GET /groups/{id}/invitations/status", authMiddleware(http.HandlerFunc(api.GetInvitationStatus)))
	mux.Handle("POST /groups/{id}/invitations/{invitationId}/{action}", authMiddleware(http.HandlerFunc(api.HandleInvitation)))

	// Group join request routes
	mux.Handle("GET /groups/{id}/join-requests", authMiddleware(http.HandlerFunc(api.GetGroupRequests)))
	mux.Handle("POST /groups/{id}/join-requests/{action}", authMiddleware(http.HandlerFunc(api.HandleJoinRequest)))
	mux.Handle("POST /groups/{id}/join", authMiddleware(http.HandlerFunc(api.RequestJoinGroup)))

	// Group events
	mux.Handle("GET /groups/{id}/events", authMiddleware(http.HandlerFunc(api.GetGroupEvents)))
	mux.Handle("POST /groups/{id}/events", authMiddleware(http.HandlerFunc(api.CreateGroupEvent)))
	mux.Handle("POST /groups/{id}/events/{eventId}/respond", authMiddleware(http.HandlerFunc(api.RespondToGroupEvent)))

	// Group posts and comments
	mux.Handle("GET /groups/{id}/posts", authMiddleware(http.HandlerFunc(api.GetGroupPost)))
	mux.Handle("POST /groups/{id}/posts", authMiddleware(http.HandlerFunc(api.CreateGroupPost)))
	mux.Handle("GET /groups/{id}/posts/{postId}/comments", authMiddleware(http.HandlerFunc(api.GetGroupPostComments)))
	mux.Handle("POST /groups/{id}/posts/{postId}/comments", authMiddleware(http.HandlerFunc(api.CreateGroupPostComment)))

	mux.Handle("POST /follow", authMiddleware(http.HandlerFunc(api.FollowUser)))
	mux.Handle("POST /unfollow", authMiddleware(http.HandlerFunc(api.UnfollowUser)))
	mux.Handle("PATCH /follow/handle-request", authMiddleware(http.HandlerFunc(api.HandleFollowRequest)))
	mux.Handle("POST /user/follow-status", authMiddleware(http.HandlerFunc(api.FollowStatus)))

	mux.Handle("GET /follower/{userID}", authMiddleware(http.HandlerFunc(api.GetFollowers)))

	mux.Handle("GET /contact/{userID}", authMiddleware(http.HandlerFunc(api.GetContact)))
	mux.Handle("GET /messages/{userId}/{contactId}", authMiddleware(http.HandlerFunc(api.GetMessages)))

	mux.Handle("GET /user/{userID}", authMiddleware(http.HandlerFunc(api.UserProfile)))
	mux.Handle("POST /updateProfile", authMiddleware(http.HandlerFunc(api.UpdateProfile)))
	mux.Handle("GET /getMyPosts", authMiddleware(http.HandlerFunc(api.GetMyPosts)))

	mux.Handle("/ws", authMiddleware(http.HandlerFunc(api.WebSocketHandler)))

	mux.Handle("GET /notifications", authMiddleware(http.HandlerFunc(api.GetNotifications)))
	mux.Handle("POST /notifications/{id}/read", authMiddleware(http.HandlerFunc(api.MarkNotificationAsRead)))

	mux.Handle("GET /uploads/group_posts/{filename}", http.HandlerFunc(api.ServeGroupPostMedia))

	mux.Handle("POST /chat/direct", http.HandlerFunc(api.CreateOrGetDirectChat))
	mux.Handle("GET /chats", http.HandlerFunc(api.GetUserChats))
	mux.Handle("GET /chat/{chatId}/participants", http.HandlerFunc(api.GetChatParticipants))
	// Group role routes
	mux.Handle("GET /groups/{id}/role", authMiddleware(http.HandlerFunc(api.GetUserRoleInGroup)))

	// Setup routes
	api.SetupRoutes(mux)

	// Wrap the entire mux with CORS middleware
	handler := middleware.CORS(mux)

	log.Println("Server running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
