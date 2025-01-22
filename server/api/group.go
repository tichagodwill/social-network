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

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID
	var creatorID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&creatorID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Parse request body
	var group struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create group
	result, err := tx.Exec(`
		INSERT INTO groups (title, description, creator_id)
		VALUES (?, ?, ?)`,
		group.Title, group.Description, creatorID)
	if err != nil {
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	// Get the new group's ID
	groupID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get new group ID", http.StatusInternalServerError)
		return
	}

	// Add creator as a member with creator role
	_, err = tx.Exec(`
		INSERT INTO group_members (group_id, user_id, role)
		VALUES (?, ?, 'creator')`,
		groupID, creatorID)
	if err != nil {
		http.Error(w, "Failed to add creator as a member", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to complete group creation", http.StatusInternalServerError)
		return
	}

	// Return success response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Group created successfully",
		"groupId": groupID,
	})
}

func CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get group ID from URL path parameter
	groupIDStr := r.PathValue("id")
	if groupIDStr == "" {
		log.Printf("Group ID is missing from URL")
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil || groupID <= 0 {
		log.Printf("Invalid group ID: %v", err)
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Verify group exists
	var exists bool
	err = sqlite.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM groups WHERE id = ?)", groupID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking group existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("Session error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("User not found: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if user is a member
	var isMember bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)`, groupID, userID).Scan(&isMember)
	if err != nil {
		log.Printf("Error checking membership: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Not a group member", http.StatusForbidden)
		return
	}

	// Parse request body
	var post struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&post); err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate post data
	if strings.TrimSpace(post.Title) == "" || strings.TrimSpace(post.Content) == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("Transaction error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create post
	result, err := tx.Exec(`
		INSERT INTO group_posts (group_id, author_id, title, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		groupID, userID, post.Title, post.Content)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	postID, _ := result.LastInsertId()

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Fetch the created post with author information
	var createdPost m.GroupPost
	err = sqlite.DB.QueryRow(`
		SELECT p.id, p.group_id, p.author_id, u.username, p.title, p.content, p.created_at, p.updated_at
		FROM group_posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.id = ?`, postID).Scan(
		&createdPost.ID,
		&createdPost.GroupID,
		&createdPost.AuthorID,
		&createdPost.Author,
		&createdPost.Title,
		&createdPost.Content,
		&createdPost.CreatedAt,
		&createdPost.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching created post: %v", err)
		http.Error(w, "Failed to fetch created post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPost)
}

func GetGroupPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if user is a member
	var isMember bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)`, groupID, userID).Scan(&isMember)
	if err != nil || !isMember {
		http.Error(w, "Not a group member", http.StatusForbidden)
		return
	}

	// Get posts with authors
	rows, err := sqlite.DB.Query(`
		SELECT p.id, p.group_id, p.author_id, u.username, p.title, p.content, p.created_at, p.updated_at
		FROM group_posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.group_id = ?
		ORDER BY p.created_at DESC`, groupID)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []m.GroupPost
	for rows.Next() {
		var post m.GroupPost
		err := rows.Scan(
			&post.ID, &post.GroupID, &post.AuthorID, &post.Author,
			&post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			continue
		}
		
		// Get comments for each post
		comments, _ := getPostComments(post.ID)
		post.Comments = comments
		posts = append(posts, post)
	}

	json.NewEncoder(w).Encode(posts)
}

func CreateGroupPostComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get group ID and post ID from URL
	groupIDStr := r.PathValue("id")
	if groupIDStr == "" {
		log.Printf("Group ID is missing from URL")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Group ID is required"})
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil || groupID <= 0 {
		log.Printf("Invalid group ID: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid group ID"})
		return
	}

	postIDStr := r.PathValue("postId")
	if postIDStr == "" {
		log.Printf("Post ID is missing from URL")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Post ID is required"})
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		log.Printf("Invalid post ID: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid post ID"})
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
		log.Printf("User not found: %v", err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	// Verify post exists and belongs to the group
	var postExists bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_posts 
			WHERE id = ? AND group_id = ?
		)`, postID, groupID).Scan(&postExists)
	if err != nil {
		log.Printf("Error checking post: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	if !postExists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Post not found"})
		return
	}

	// Check if user is a member
	var isMember bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)`, groupID, userID).Scan(&isMember)
	if err != nil {
		log.Printf("Error checking membership: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	if !isMember {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not a group member"})
		return
	}

	// Parse request body
	var comment struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		log.Printf("Invalid request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate comment content
	if strings.TrimSpace(comment.Content) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Comment content is required"})
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("Transaction error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Create comment
	result, err := tx.Exec(`
		INSERT INTO group_post_comments (post_id, author_id, content, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		postID, userID, comment.Content)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create comment"})
		return
	}

	commentID, _ := result.LastInsertId()

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create comment"})
		return
	}

	// Fetch the created comment with author information
	var createdComment m.GroupPostComment
	err = sqlite.DB.QueryRow(`
		SELECT c.id, c.post_id, c.author_id, u.username, c.content, c.created_at, c.updated_at
		FROM group_post_comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.id = ?`, commentID).Scan(
		&createdComment.ID,
		&createdComment.PostID,
		&createdComment.AuthorID,
		&createdComment.Author,
		&createdComment.Content,
		&createdComment.CreatedAt,
		&createdComment.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching created comment: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch created comment"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdComment)
}

func getPostComments(postID int) ([]m.GroupPostComment, error) {
	rows, err := sqlite.DB.Query(`
		SELECT c.id, c.post_id, c.author_id, u.username, c.content, c.created_at, c.updated_at
		FROM group_post_comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []m.GroupPostComment
	for rows.Next() {
		var comment m.GroupPostComment
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.AuthorID, &comment.Author,
			&comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			continue
		}
		comments = append(comments, comment)
	}
	return comments, nil
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

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

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
		if err == sql.ErrNoRows {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if user is a member or creator
	var isMember bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)`, groupID, userID).Scan(&isMember)
	
	if err != nil {
		log.Printf("Error checking membership: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":              group.ID,
		"title":          group.Title,
		"description":    group.Description,
		"creator_id":     group.CreatorID,
		"creator_username": group.CreatorUsername,
		"created_at":     group.CreatedAt,
		"is_member":      isMember,
	}

	json.NewEncoder(w).Encode(response)
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
		GroupID        int    `json:"groupId"`
		Identifier    string `json:"identifier"`
		IdentifierType string `json:"identifierType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request format",
		})
		return
	}

	// Get current user (inviter)
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONResponse(w, http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized",
		})
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Database error",
		})
		return
	}
	defer tx.Rollback()

	// Get inviter ID and check if they're the creator
	var inviterID int
	var isCreator bool
	err = tx.QueryRow(`
		SELECT u.id, g.creator_id = u.id
		FROM users u
		JOIN groups g ON g.id = ?
		WHERE u.username = ?`,
		req.GroupID, username).Scan(&inviterID, &isCreator)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJSONResponse(w, http.StatusNotFound, map[string]interface{}{
				"error": "Group not found",
			})
			return
		}
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to verify permissions",
		})
		return
	}

	if !isCreator {
		sendJSONResponse(w, http.StatusForbidden, map[string]interface{}{
			"error": "Only group creator can invite members",
		})
		return
	}

	// Get invitee information
	var inviteeID int
	var inviteeUsername string
	var query string
	if req.IdentifierType == "email" {
		query = "SELECT id, username FROM users WHERE email = ?"
	} else {
		query = "SELECT id, username FROM users WHERE username = ?"
	}

	err = tx.QueryRow(query, req.Identifier).Scan(&inviteeID, &inviteeUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJSONResponse(w, http.StatusNotFound, map[string]interface{}{
				"error": "User not found",
			})
			return
		}
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to find user",
		})
		return
	}

	// Check if user is already a member
	var isMember bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)`, req.GroupID, inviteeID).Scan(&isMember)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to check membership",
		})
		return
	}
	if isMember {
		sendJSONResponse(w, http.StatusBadRequest, map[string]interface{}{
			"error": "User is already a member",
		})
		return
	}

	// Check for existing invitation
	var hasPendingInvitation bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_invitations 
			WHERE group_id = ? AND invitee_id = ? AND status = 'pending'
		)`, req.GroupID, inviteeID).Scan(&hasPendingInvitation)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to check existing invitations",
		})
		return
	}
	if hasPendingInvitation {
		sendJSONResponse(w, http.StatusBadRequest, map[string]interface{}{
			"error": "User already has a pending invitation",
		})
		return
	}

	// Create invitation
	result, err := tx.Exec(`
		INSERT INTO group_invitations (group_id, inviter_id, invitee_id, status)
		VALUES (?, ?, ?, 'pending')`,
		req.GroupID, inviterID, inviteeID)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create invitation",
		})
		return
	}

	invitationID, err := result.LastInsertId()
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to process invitation",
		})
		return
	}

	// Create notification
	_, err = tx.Exec(`
		INSERT INTO notifications (user_id, type, content, from_user_id, group_id)
		VALUES (?, 'group_invitation', ?, ?, ?)`,
		inviteeID, fmt.Sprintf("%s invited you to join their group", username), inviterID, req.GroupID)
	if err != nil {
		log.Printf("Create notification error: %v", err)
	}

	if err = tx.Commit(); err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to complete invitation",
		})
		return
	}

	// Send success response with the invitation ID
	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Invitation sent successfully",
		"invitationId": invitationID,
		"invitedUser": inviteeUsername,
	})
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
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Check if current user is the group creator
	isCreator, err := checkUserRole(groupID, userID, "creator")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !isCreator {
		http.Error(w, "Only group creator can remove members", http.StatusForbidden)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check if target user is not the creator
	var targetRole string
	err = tx.QueryRow(`
		SELECT role FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, memberID).Scan(&targetRole)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Member not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if targetRole == "creator" {
		http.Error(w, "Cannot remove group creator", http.StatusForbidden)
		return
	}

	// Remove member
	_, err = tx.Exec(`
		DELETE FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, memberID)
	if err != nil {
		http.Error(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	// Create notification for removed member
	_, err = tx.Exec(`
		INSERT INTO notifications (user_id, type, content, from_user_id, group_id)
		VALUES (?, 'group_removal', 'You have been removed from the group', ?, ?)`,
		memberID, userID, groupID)
	if err != nil {
		log.Printf("Failed to create notification: %v", err)
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to complete member removal", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member removed successfully",
	})
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

// GetInvitationStatus checks if a user has a pending invitation or request for a group
func GetInvitationStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid group ID",
		})
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONResponse(w, http.StatusUnauthorized, map[string]interface{}{
			"error": "Unauthorized",
		})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get user information",
		})
		return
	}

	// Check for invitation (where inviter_id != invitee_id)
	var invitation struct {
		ID        int    `json:"id"`
		Status    string `json:"status"`
		CreatedAt string `json:"createdAt"`
	}
	var hasInvitation bool
	err = sqlite.DB.QueryRow(`
		SELECT id, status, created_at 
		FROM group_invitations 
		WHERE group_id = ? AND invitee_id = ? AND inviter_id != invitee_id AND status = 'pending'`,
		groupID, userID).Scan(&invitation.ID, &invitation.Status, &invitation.CreatedAt)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Invitation check error: %v", err)
			sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
				"error": "Failed to check invitation status",
			})
			return
		}
		hasInvitation = false
	} else {
		hasInvitation = true
	}

	// Check for join request (where inviter_id = invitee_id)
	var hasRequest bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_invitations 
			WHERE group_id = ? AND invitee_id = ? AND inviter_id = invitee_id AND status = 'pending'
		)`, groupID, userID).Scan(&hasRequest)
	if err != nil {
		log.Printf("Request check error: %v", err)
		sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to check request status",
		})
		return
	}

	response := map[string]interface{}{
		"hasInvitation": hasInvitation,
		"hasRequest":    hasRequest,
	}
	
	if hasInvitation {
		response["invitation"] = invitation
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// RequestJoinGroup handles requests to join a group
func RequestJoinGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
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
		log.Printf("User lookup error: %v", err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Check if group exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM groups WHERE id = ?)", groupID).Scan(&exists)
	if err != nil {
		log.Printf("Group existence check error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check group"})
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
		log.Printf("Member check error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
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
		log.Printf("Invitation check error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check existing invitations"})
		return
	}
	if exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Already have a pending request or invitation"})
		return
	}

	// Create the join request
	result, err := tx.Exec(`
		INSERT INTO group_invitations (group_id, inviter_id, invitee_id, status)
		VALUES (?, ?, ?, 'pending')`,
		groupID, userID, userID) // inviter_id = invitee_id indicates it's a request
	if err != nil {
		log.Printf("Insert invitation error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create join request"})
		return
	}

	// Get the invitation ID
	invitationID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Get last insert ID error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get invitation ID"})
		return
	}

	// Get group creator for notification
	var creatorID int
	err = tx.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
	if err != nil {
		log.Printf("Get creator ID error: %v", err)
	} else {
		// Create notification for group creator
		_, err = tx.Exec(`
			INSERT INTO notifications (user_id, type, content, from_user_id, group_id)
			VALUES (?, 'group_join_request', ?, ?, ?)`,
			creatorID, fmt.Sprintf("%s requested to join your group", username), userID, groupID)
		if err != nil {
			log.Printf("Create notification error: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Transaction commit error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to complete join request"})
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
	
	invitationID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid invitation ID"})
		return
	}
	
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
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Get invitation details
	var invitation struct {
		GroupID   int
		InviterID int
		InviteeID int
		Status    string
	}
	err = tx.QueryRow(`
		SELECT group_id, inviter_id, invitee_id, status 
		FROM group_invitations 
		WHERE id = ? AND status = 'pending'`,
		invitationID).Scan(&invitation.GroupID, &invitation.InviterID, &invitation.InviteeID, &invitation.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invitation not found or already processed"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		}
		return
	}

	if action == "accept" {
		// Add member to group
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, user_id, role, status)
			VALUES (?, ?, 'member', 'active')`,
			invitation.GroupID, invitation.InviteeID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add member to group"})
			return
		}
	}

	// Update invitation status
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

	// Delete related notifications
	_, err = tx.Exec(`
		DELETE FROM notifications 
		WHERE (type = 'group_join_request' OR type = 'group_invitation') 
		AND group_id = ? 
		AND (user_id = ? OR from_user_id = ?)`,
		invitation.GroupID, invitation.InviteeID, invitation.InviteeID)
	if err != nil {
		log.Printf("Failed to delete notifications: %v", err)
	}

	if err = tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to complete action"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Invitation %sed successfully", action),
	})
}

// Helper function for consistent JSON responses
func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
