package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

// Helper function to check if user has required role
func checkUserRole(groupID, userID int, requiredRole string) (bool, error) {
	var role string
	err := sqlite.DB.QueryRow(`
		SELECT role 
		FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, userID).Scan(&role)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// For creator role, only exact match
	if requiredRole == "creator" {
		return role == "creator", nil
	}

	// Role hierarchy: creator > admin > moderator > member
	roleHierarchy := map[string]int{
		"creator":   4,
		"admin":     3,
		"moderator": 2,
		"member":    1,
	}

	return roleHierarchy[role] >= roleHierarchy[requiredRole], nil
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the current user from the session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Get user ID
	var creatorID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&creatorID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	// Parse the request body
	var group struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON format"})
		return
	}

	// Validate required fields
	if strings.TrimSpace(group.Title) == "" || strings.TrimSpace(group.Description) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Please provide all required fields"})
		return
	}

	// Start a transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback() // Will rollback if not committed

	// Insert the group
	result, err := tx.Exec(
		"INSERT INTO groups (creator_id, title, description, created_at) VALUES (?, ?, ?, datetime('now'))",
		creatorID, group.Title, group.Description)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create group"})
		log.Printf("Error creating group: %v", err)
		return
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve group ID"})
		return
	}

	// Add creator as a member with "creator" role and "active" status
	_, err = tx.Exec(
		"INSERT INTO group_members (group_id, user_id, role, status) VALUES (?, ?, ?, ?)",
		groupID, creatorID, "creator", "active")
	if err != nil {
		log.Printf("Error adding creator as member: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add creator as a member"})
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to commit transaction"})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      groupID,
		"message": "Group created successfully",
	})
}

func CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the group ID from URL
	groupIDStr := r.URL.Query().Get("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// Parse the request body
	var post m.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON format"})
		return
	}

	// Set post privacy to public by default for group posts
	post.Privacy = 1

	// Validate post content
	if strings.TrimSpace(post.Title) == "" || strings.TrimSpace(post.Content) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Title and content cannot be empty"})
		return
	}

	// Insert the post into the database
	_, err = sqlite.DB.Exec(
		"INSERT INTO posts (title, content, media, privacy, author, group_id) VALUES (?, ?, ?, ?, ?, ?)",
		post.Title, post.Content, post.Media, post.Privacy, post.Author, groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create post"})
		log.Printf("Error creating post: %v", err)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Post created successfully"})
}

func GetGroupPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the group ID from URL
	groupIDStr := r.URL.Query().Get("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil || groupID < 1 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// Fetch posts from the database
	rows, err := sqlite.DB.Query(
		"SELECT id, title, content, media, privacy, author, created_at, group_id FROM posts WHERE group_id = ?",
		groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch posts"})
		log.Printf("Error fetching posts: %v", err)
		return
	}
	defer rows.Close()

	// Iterate through the results
	var posts []m.Post
	for rows.Next() {
		var post m.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privacy, &post.Author, &post.CreatedAt, &post.GroupID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error reading posts"})
			log.Printf("Error scanning post: %v", err)
			return
		}
		posts = append(posts, post)
	}

	// Return the posts as JSON
	json.NewEncoder(w).Encode(posts)
}

func ViewGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	// Fetch groups
	rows, err := sqlite.DB.Query(`
		SELECT g.id, g.title, g.description, g.creator_id, g.created_at,
			   u.username as creator_username,
			   EXISTS(SELECT 1 FROM group_members WHERE group_id = g.id AND user_id = ?) as is_member,
			   EXISTS(SELECT 1 FROM group_invitations WHERE group_id = g.id AND invitee_id = ? AND status = 'pending') as has_pending_request
		FROM groups g
		JOIN users u ON g.creator_id = u.id`, userID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch groups"})
		return
	}
	defer rows.Close()

	var groups []map[string]interface{}
	for rows.Next() {
		var group struct {
			ID              int       `json:"id"`
			Title           string    `json:"title"`
			Description     string    `json:"description"`
			CreatorID       int       `json:"creator_id"`
			CreatedAt       time.Time `json:"created_at"`
			CreatorUsername string    `json:"creator_username"`
			IsMember        bool      `json:"is_member"`
			HasPendingRequest bool    `json:"has_pending_request"`
		}

		err := rows.Scan(
			&group.ID, &group.Title, &group.Description, &group.CreatorID,
			&group.CreatedAt, &group.CreatorUsername, &group.IsMember, &group.HasPendingRequest)
		if err != nil {
			continue
		}

		groups = append(groups, map[string]interface{}{
			"id":                group.ID,
			"title":            group.Title,
			"description":      group.Description,
			"creator_id":       group.CreatorID,
			"created_at":       group.CreatedAt,
			"creator_username": group.CreatorUsername,
			"is_member":        group.IsMember,
			"has_pending_request": group.HasPendingRequest,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(groups)
}

func GetGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get group ID from URL
	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		log.Printf("Invalid group ID: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("Session error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Database error getting user ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	// Get group information with error handling
	var group m.Group
	err = sqlite.DB.QueryRow(`
		SELECT 
			g.id, 
			g.title, 
			g.description, 
			g.creator_id, 
			u.username as creator_username,
			g.created_at
		FROM groups g
		JOIN users u ON g.creator_id = u.id
		WHERE g.id = ?`, groupID).Scan(
		&group.ID,
		&group.Title,
		&group.Description,
		&group.CreatorID,
		&group.CreatorUsername,
		&group.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Group not found: %d", groupID)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}
		log.Printf("Database error getting group: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get group information"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

func GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	rows, err := sqlite.DB.Query(`
		SELECT u.id, u.username, gm.role
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ?`, groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch members"})
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var member struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		}
		if err := rows.Scan(&member.ID, &member.Username, &member.Role); err != nil {
			continue
		}
		members = append(members, map[string]interface{}{
			"id":       member.ID,
			"username": member.Username,
			"role":     member.Role,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(members)
}

func VeiwGorups(w http.ResponseWriter, r *http.Request) {
	ViewGroups(w, r)
}

func GetGroupPost(w http.ResponseWriter, r *http.Request) {
	GetGroupPosts(w, r)
}

func GroupInvitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req struct {
		GroupID       int    `json:"groupId"`
		Identifier   string `json:"identifier"`
		IdentifierType string `json:"identifierType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}

	// Get inviter's username from session
	inviterUsername, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Get inviter's ID
	var inviterID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", inviterUsername).Scan(&inviterID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get inviter information"})
		return
	}

	// Find invitee by email or username
	var inviteeID int
	var query string
	if req.IdentifierType == "email" {
		query = "SELECT id FROM users WHERE email = ?"
	} else {
		query = "SELECT id FROM users WHERE username = ?"
	}

	err = sqlite.DB.QueryRow(query, req.Identifier).Scan(&inviteeID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}

	// Check if user is already a member
	var exists bool
	err = sqlite.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)",
		req.GroupID, inviteeID).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	if exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User is already a member"})
		return
	}

	// Create invitation
	_, err = sqlite.DB.Exec(`
		INSERT INTO group_invitations (group_id, inviter_id, invitee_id, status)
		VALUES (?, ?, ?, 'pending')
		ON CONFLICT (group_id, invitee_id) DO UPDATE SET
		status = 'pending', inviter_id = ?`,
		req.GroupID, inviterID, inviteeID, inviterID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create invitation"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Invitation sent successfully"})
}

func GroupAccept(w http.ResponseWriter, r *http.Request) {
	// Implementation for accepting group invitation
}

func GroupReject(w http.ResponseWriter, r *http.Request) {
	// Implementation for rejecting group invitation
}

func GroupLeave(w http.ResponseWriter, r *http.Request) {
	// Implementation for leaving a group
}

func GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// First check if the group exists
	var exists bool
	err = sqlite.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM groups WHERE id = ?)", groupID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking group existence: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check group"})
		return
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
		return
	}

	// Get events with response counts
	rows, err := sqlite.DB.Query(`
		SELECT 
			e.id,
			e.title,
			e.description,
			e.event_date,
			e.creator_id,
			COALESCE(
				(SELECT COUNT(*) FROM event_responses 
				WHERE event_id = e.id AND status = 'going'), 0
			) as going_count,
			COALESCE(
				(SELECT COUNT(*) FROM event_responses 
				WHERE event_id = e.id AND status = 'not_going'), 0
			) as not_going_count
		FROM group_events e
		WHERE e.group_id = ?
		ORDER BY e.event_date DESC`, groupID)
	if err != nil {
		log.Printf("Error fetching events: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch events"})
		return
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var event struct {
			ID            int       `json:"id"`
			Title         string    `json:"title"`
			Description   string    `json:"description"`
			EventDate     time.Time `json:"eventDate"`
			CreatorID     int       `json:"creatorId"`
			GoingCount    int       `json:"goingCount"`
			NotGoingCount int       `json:"notGoingCount"`
		}

		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&event.CreatorID,
			&event.GoingCount,
			&event.NotGoingCount,
		); err != nil {
			log.Printf("Error scanning event row: %v", err)
			continue
		}

		events = append(events, map[string]interface{}{
			"id":            event.ID,
			"title":         event.Title,
			"description":   event.Description,
			"eventDate":     event.EventDate.Format(time.RFC3339),
			"creatorId":     event.CreatorID,
			"goingCount":    event.GoingCount,
			"notGoingCount": event.NotGoingCount,
		})
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating events: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error processing events"})
		return
	}

	// If no events found, return empty array instead of null
	if events == nil {
		events = make([]map[string]interface{}, 0)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}

func CreateGroupEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// Parse request body
	var event struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		EventDate   string    `json:"eventDate"`
		CreatorID   int       `json:"creatorId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Parse event date
	eventDate, err := time.Parse(time.RFC3339, event.EventDate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid date format"})
		return
	}

	// Insert the event
	result, err := sqlite.DB.Exec(`
		INSERT INTO group_events (
			group_id, title, description, event_date, creator_id
		) VALUES (?, ?, ?, ?, ?)`,
		groupID, event.Title, event.Description, eventDate, event.CreatorID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create event"})
		log.Printf("Error creating event: %v", err)
		return
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get event ID"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": eventID,
		"message": "Event created successfully",
	})
}

func RespondToGroupEvent(w http.ResponseWriter, r *http.Request) {
	// Implementation for responding to a group event
}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// Parse request body
	var updateData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	// Check if user is creator
	var creatorID int
	err = sqlite.DB.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check group"})
		return
	}

	if creatorID != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only the creator can update this group"})
		return
	}

	// Update the group
	_, err = sqlite.DB.Exec(`
		UPDATE groups 
		SET title = ?, description = ? 
		WHERE id = ?`,
		updateData.Title, updateData.Description, groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update group"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group updated successfully"})
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	// Check if user is creator
	var creatorID int
	err = sqlite.DB.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check group"})
		return
	}

	if creatorID != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only the creator can delete this group"})
		return
	}

	// Delete the group
	_, err = sqlite.DB.Exec("DELETE FROM groups WHERE id = ?", groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete group"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group deleted successfully"})
}

func UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	memberIDStr := r.PathValue("memberId")

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid member ID"})
		return
	}

	// Parse request body
	var updateData struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate role
	validRoles := map[string]bool{"member": true, "moderator": true, "admin": true}
	if !validRoles[updateData.Role] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid role"})
		return
	}

	// Verify user has permission (creator or admin)
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	// Check if user is creator
	var role string
	err = sqlite.DB.QueryRow(`
		SELECT role FROM group_members 
		WHERE group_id = ? AND user_id = ? AND role = 'creator'`,
		groupID, userID).Scan(&role)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only the creator can update member roles"})
		return
	}

	// Update member role
	result, err := sqlite.DB.Exec(`
		UPDATE group_members 
		SET role = ? 
		WHERE group_id = ? AND user_id = ? AND role != 'creator'`,
		updateData.Role, groupID, memberID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update member role"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Member not found or is the creator"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member role updated successfully"})
}

func RemoveMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	memberIDStr := r.PathValue("memberId")

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid member ID"})
		return
	}

	// Verify user has permission (creator or admin)
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	// Check if user is creator
	var role string
	err = sqlite.DB.QueryRow(`
		SELECT role FROM group_members 
		WHERE group_id = ? AND user_id = ? AND role = 'creator'`,
		groupID, userID).Scan(&role)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only the creator can remove members"})
		return
	}

	// Check if target member is not the creator
	var memberRole string
	err = sqlite.DB.QueryRow(`
		SELECT role FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, memberID).Scan(&memberRole)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Member not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check member role"})
		return
	}

	if memberRole == "creator" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot remove the group creator"})
		return
	}

	// Remove the member
	result, err := sqlite.DB.Exec(`
		DELETE FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, memberID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to remove member"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Member not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member removed successfully"})
}

func GetGroupRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	rows, err := sqlite.DB.Query(`
		SELECT gi.id, gi.invitee_id, u.username, gi.created_at
		FROM group_invitations gi
		JOIN users u ON gi.invitee_id = u.id
		WHERE gi.group_id = ? AND gi.status = 'pending'`, groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch requests"})
		return
	}
	defer rows.Close()

	var requests []map[string]interface{}
	for rows.Next() {
		var request struct {
			ID        int       `json:"id"`
			InviteeID int       `json:"invitee_id"`
			Username  string    `json:"username"`
			CreatedAt time.Time `json:"created_at"`
		}
		if err := rows.Scan(&request.ID, &request.InviteeID, &request.Username, &request.CreatedAt); err != nil {
			continue
		}
		requests = append(requests, map[string]interface{}{
			"id":         request.ID,
			"inviteeId":  request.InviteeID,
			"username":   request.Username,
			"createdAt":  request.CreatedAt,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(requests)
}

// GetInvitationStatus checks if a user has a pending invitation or join request
func GetInvitationStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	// Check for pending invitation
	var invitation struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}
	err = sqlite.DB.QueryRow(`
		SELECT id, status 
		FROM group_invitations 
		WHERE group_id = ? AND invitee_id = ? AND status = 'pending'`,
		groupID, userID).Scan(&invitation.ID, &invitation.Status)

	hasInvitation := err != sql.ErrNoRows
	if err != nil && err != sql.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check invitation status"})
		return
	}

	// Check for pending join request
	var hasRequest bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_invitations 
			WHERE group_id = ? AND inviter_id = ? AND invitee_id = ? AND status = 'pending'
		)`, groupID, userID, userID).Scan(&hasRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check request status"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hasInvitation": hasInvitation,
		"hasRequest":    hasRequest,
		"invitation":    hasInvitation ? invitation : nil,
	})
}

// RequestJoinGroup handles requests to join a group
func RequestJoinGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Check if group exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM groups WHERE id = ?)", groupID).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check group existence"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Group not found"})
		return
	}

	// Check if user is already a member
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)`, groupID, userID).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check membership"})
		return
	}
	if exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Already a member of this group"})
		return
	}

	// Check for existing invitation or request
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_invitations 
			WHERE group_id = ? AND invitee_id = ? AND status = 'pending'
		)`, groupID, userID).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check existing requests"})
		return
	}
	if exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Already have a pending request or invitation"})
		return
	}

	// Create the join request
	result, err := tx.Exec(`
		INSERT INTO group_invitations (group_id, inviter_id, invitee_id, status, created_at)
		VALUES (?, ?, ?, 'pending', CURRENT_TIMESTAMP)`,
		groupID, userID, userID) // inviter_id = invitee_id indicates it's a request
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create join request"})
		return
	}

	// Get the invitation ID
	invitationID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get invitation ID"})
		return
	}

	// Get group creator for notification
	var creatorID int
	err = tx.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
	if err == nil {
		// Create notification for group creator
		_, err = tx.Exec(`
			INSERT INTO notifications (user_id, type, content, from_user_id, group_id)
			VALUES (?, 'group_join_request', ?, ?, ?)`,
			creatorID, fmt.Sprintf("%s requested to join your group", username), userID, groupID)
		if err != nil {
			log.Printf("Failed to create notification: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to commit transaction"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Join request sent successfully",
		"invitationId": invitationID,
	})
}

// HandleInvitation handles accepting or rejecting an invitation
func HandleInvitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	invitationID := r.PathValue("id")
	action := r.PathValue("action")
	
	if action != "accept" && action != "reject" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid action"})
		return
	}

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get user information"})
		return
	}

	tx, err := sqlite.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	var groupID int
	err = tx.QueryRow(`
		SELECT group_id FROM group_invitations 
		WHERE id = ? AND invitee_id = ? AND status = 'pending'`,
		invitationID, userID).Scan(&groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invitation not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get invitation"})
		}
		return
	}

	if action == "accept" {
		// Add member to group
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, user_id, role, status)
			VALUES (?, ?, 'member', 'active')`,
				groupID, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add member"})
			return
		}

		// Create notification for the new member
		_, err = tx.Exec(`
			INSERT INTO notifications (user_id, type, content, group_id)
			VALUES (?, 'group_join_accepted', 'You are now a member of the group', ?)`,
			userID, groupID)
		if err != nil {
			log.Printf("Failed to create member notification: %v", err)
		}

		// Create notification for group creator
		var creatorID int
		err = tx.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
		if err == nil {
			_, err = tx.Exec(`
				INSERT INTO notifications (user_id, type, content, from_user_id, group_id)
				VALUES (?, 'group_join', ?, ?, ?)`,
				creatorID, fmt.Sprintf("%s joined your group", username), userID, groupID)
			if err != nil {
				log.Printf("Failed to create creator notification: %v", err)
			}
		}
	}

	_, err = tx.Exec(`
		UPDATE group_invitations 
		SET status = ? 
			WHERE id = ?`,
		action+"ed", invitationID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update invitation"})
		return
	}

	if err = tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to commit transaction"})
		return
	}

	// Return updated group data
	var group m.Group
	err = sqlite.DB.QueryRow(`
		SELECT g.id, g.title, g.description, g.creator_id, u.username as creator_username, g.created_at
		FROM groups g
		JOIN users u ON g.creator_id = u.id
		WHERE g.id = ?`, groupID).Scan(
		&group.ID,
		&group.Title,
		&group.Description,
		&group.CreatorID,
		&group.CreatorUsername,
		&group.CreatedAt)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get updated group data"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Invitation " + action + "ed successfully",
		"group": group,
	})
}
