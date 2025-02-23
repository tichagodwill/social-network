package api

import (
	"database/sql"
	"log"
	"net/http"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

// GetUserRoleInGroup handles getting user role in a group
func GetUserRoleInGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get group ID from URL
	groupID := r.PathValue("id")
	if groupID == "" {
		sendJSONError(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Get username from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's role in the group
	var role string
	err = sqlite.DB.QueryRow(`
        SELECT role 
        FROM group_members 
        WHERE group_id = ? 
        AND user_id = (SELECT id FROM users WHERE username = ?)`,
		groupID, username).Scan(&role)

	if err == sql.ErrNoRows {
		// User is not a member of the group
		sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"role": nil,
		})
		return
	}

	if err != nil {
		log.Printf("Error getting user role: %v", err)
		sendJSONError(w, "Failed to get user role", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"role": role,
	})
} 