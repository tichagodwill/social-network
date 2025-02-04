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

// Helper function to check if user has admin privileges
func hasAdminPrivileges(groupID int, userID int) (bool, error) {
	var isCreatorOrAdmin bool
	err := sqlite.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM group_members gm
			WHERE gm.group_id = ? 
			AND gm.user_id = ? 
			AND (gm.role = 'creator' OR gm.role = 'admin')
		)`, groupID, userID).Scan(&isCreatorOrAdmin)

	if err != nil {
		log.Printf("Error checking admin privileges: %v", err)
		return false, err
	}

	return isCreatorOrAdmin, nil
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

	// Get posts with authors and comments
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

	// Parse path parameters
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.PathValue("postId"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
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

	// Verify post exists and belongs to the group
	var postExists bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_posts 
			WHERE id = ? AND group_id = ?
		)`, postID, groupID).Scan(&postExists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !postExists {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Parse request body
	var commentData struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&commentData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate content
	if strings.TrimSpace(commentData.Content) == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create comment
	result, err := tx.Exec(`
		INSERT INTO group_post_comments (post_id, author_id, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		postID, userID, commentData.Content)
	if err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	commentID, _ := result.LastInsertId()

	// Get the created comment with author info
	var comment struct {
		ID        int64  `json:"id"`
		Content   string `json:"content"`
		AuthorID  int64  `json:"author_id"`
		Author    string `json:"author"`
		CreatedAt string `json:"created_at"`
	}

	err = tx.QueryRow(`
		SELECT c.id, c.content, c.author_id, u.username, c.created_at
		FROM group_post_comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.id = ?`,
		commentID).Scan(&comment.ID, &comment.Content, &comment.AuthorID, &comment.Author, &comment.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to fetch created comment", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to complete comment creation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
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
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Fetch groups with membership status
	rows, err := sqlite.DB.Query(`
		SELECT 
			g.id, 
			g.title, 
			g.description, 
			g.creator_id, 
			u.username as creator_username,
			g.created_at,
			CASE WHEN gm.user_id IS NOT NULL THEN 1 ELSE 0 END as is_member
		FROM groups g
		JOIN users u ON g.creator_id = u.id
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.user_id = ?
		ORDER BY g.created_at DESC
	`, userID)

	if err != nil {
		log.Printf("Database error: %v", err)
		sendJSONError(w, "Failed to fetch groups", http.StatusInternalServerError)
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
			CreatorUsername string    `json:"creator_username"`
			CreatedAt       time.Time `json:"created_at"`
			IsMember        bool      `json:"is_member"`
		}

		err := rows.Scan(
			&group.ID,
			&group.Title,
			&group.Description,
			&group.CreatorID,
			&group.CreatorUsername,
			&group.CreatedAt,
			&group.IsMember,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		groups = append(groups, map[string]interface{}{
			"id":               group.ID,
			"title":            group.Title,
			"description":      group.Description,
			"creator_id":       group.CreatorID,
			"creator_username": group.CreatorUsername,
			"created_at":       group.CreatedAt,
			"is_member":        group.IsMember,
		})
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		sendJSONError(w, "Error processing groups", http.StatusInternalServerError)
		return
	}

	// If no groups found, return empty array instead of null
	if groups == nil {
		groups = make([]map[string]interface{}, 0)
	}

	sendJSONResponse(w, http.StatusOK, groups)
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
		"id":               group.ID,
		"title":            group.Title,
		"description":      group.Description,
		"creator_id":       group.CreatorID,
		"creator_username": group.CreatorUsername,
		"created_at":       group.CreatedAt,
		"is_member":        isMember,
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

func InviteToGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		sendJSONError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var inviterID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&inviterID)
	if err != nil {
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Parse request body
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get invitee ID
	var inviteeID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&inviteeID)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJSONError(w, "User not found", http.StatusNotFound)
		} else {
			sendJSONError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check if user is already a member
	var isMember bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)`, groupID, inviteeID).Scan(&isMember)
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if isMember {
		sendJSONError(w, "User is already a member", http.StatusBadRequest)
		return
	}

	// Check for existing invitation
	var hasInvitation bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_invitations 
			WHERE group_id = ? AND invitee_id = ? AND status = 'pending'
		)`, groupID, inviteeID).Scan(&hasInvitation)
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if hasInvitation {
		sendJSONError(w, "User already has a pending invitation", http.StatusBadRequest)
		return
	}

	// Get group information
	var group struct {
		Title string
	}
	err = tx.QueryRow(`SELECT title FROM groups WHERE id = ?`, groupID).Scan(&group.Title)
	if err != nil {
		sendJSONError(w, "Failed to get group information", http.StatusInternalServerError)
		return
	}

	// Create invitation
	result, err := tx.Exec(`
		INSERT INTO group_invitations (
			group_id, 
			inviter_id, 
			invitee_id, 
			type,
			status, 
			created_at
		) VALUES (?, ?, ?, 'invitation', 'pending', CURRENT_TIMESTAMP)`,
		groupID, inviterID, inviteeID)
	if err != nil {
		sendJSONError(w, "Failed to create invitation", http.StatusInternalServerError)
		return
	}

	invitationID, err := result.LastInsertId()
	if err != nil {
		sendJSONError(w, "Failed to get invitation ID", http.StatusInternalServerError)
		return
	}

	// Add debug logging
	log.Printf("Created invitation with ID: %d", invitationID)

	// Create notification with the invitation ID
	_, err = tx.Exec(`
		INSERT INTO notifications (
			user_id,
			type,
			content,
			group_id,
			invitation_id,
			from_user_id,
			is_read,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, false, CURRENT_TIMESTAMP)`,
		inviteeID,
		"group_invitation",
		fmt.Sprintf("%s has invited you to join %s", username, group.Title),
		groupID,
		invitationID,
		inviterID)

	// Add debug logging
	log.Printf("Created notification for invitation ID: %d", invitationID)

	if err != nil {
		sendJSONError(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	// Send WebSocket notification
	if err == nil {
		notification := map[string]interface{}{
			"type": "notification",
			"data": map[string]interface{}{
				"id":           invitationID,
				"type":         "group_invitation",
				"content":      fmt.Sprintf("%s has invited you to join %s", username, group.Title),
				"groupId":      groupID,
				"invitationId": invitationID,
				"fromUserId":   inviterID,
				"userId":       inviteeID,
				"createdAt":    time.Now().Format(time.RFC3339),
				"isRead":       false,
				"isProcessed":  false,
			},
		}

		// Send to specific user (the invitee)
		SendNotification([]uint64{uint64(inviteeID)}, notification)
	}

	if err = tx.Commit(); err != nil {
		sendJSONError(w, "Failed to complete invitation", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Invitation sent successfully",
	})
}

// Helper function to send JSON errors
func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// Helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
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
		Title       string `json:"title"`
		Description string `json:"description"`
		EventDate   string `json:"eventDate"`
		CreatorID   int    `json:"creatorId"`
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
		"id":      eventID,
		"message": "Event created successfully",
	})
}

func RespondToGroupEvent(w http.ResponseWriter, r *http.Request) {
	// Implementation for responding to a group event
}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
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

	// Verify user is creator
	var isCreator bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM groups 
			WHERE id = ? AND creator_id = (SELECT id FROM users WHERE username = ?)
		)`, groupID, username).Scan(&isCreator)
	if err != nil || !isCreator {
		http.Error(w, "Only group creator can update group", http.StatusForbidden)
		return
	}

	var updateData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = sqlite.DB.Exec(`
		UPDATE groups 
		SET title = ?, description = ? 
		WHERE id = ?`,
		updateData.Title, updateData.Description, groupID)
	if err != nil {
		http.Error(w, "Failed to update group", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Group updated successfully",
	})
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
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

	// Verify user is creator
	var isCreator bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM groups 
			WHERE id = ? AND creator_id = (SELECT id FROM users WHERE username = ?)
		)`, groupID, username).Scan(&isCreator)
	if err != nil || !isCreator {
		http.Error(w, "Only group creator can delete group", http.StatusForbidden)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete group and all related data
	_, err = tx.Exec("DELETE FROM groups WHERE id = ?", groupID)
	if err != nil {
		http.Error(w, "Failed to delete group", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to complete deletion", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Group deleted successfully",
	})
}

func UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if method is PUT
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	memberID, err := strconv.Atoi(r.PathValue("memberId"))
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var updateData struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role
	validRoles := map[string]bool{
		"member":    true,
		"moderator": true,
		"admin":     true,
	}
	if !validRoles[updateData.Role] {
		http.Error(w, "Invalid role", http.StatusBadRequest)
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
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Check if user has admin privileges
	hasPrivileges, err := hasAdminPrivileges(groupID, userID)
	if err != nil {
		http.Error(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !hasPrivileges {
		http.Error(w, "Only group creator and admins can update member roles", http.StatusForbidden)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get current role of the user making the change
	var currentUserRole string
	err = tx.QueryRow(`
		SELECT role FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, userID).Scan(&currentUserRole)
	if err != nil {
		http.Error(w, "Failed to get user role", http.StatusInternalServerError)
		return
	}

	// Get target member's current role
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

	// Role hierarchy
	roleHierarchy := map[string]int{
		"creator":   4,
		"admin":     3,
		"moderator": 2,
		"member":    1,
	}

	// Only creator can modify admin roles
	if targetRole == "creator" || (currentUserRole != "creator" && roleHierarchy[targetRole] >= roleHierarchy[currentUserRole]) {
		http.Error(w, "Insufficient permissions to modify this role", http.StatusForbidden)
		return
	}

	// Update member role
	result, err := tx.Exec(`
		UPDATE group_members 
		SET role = ? 
		WHERE group_id = ? AND user_id = ? AND role != 'creator'`,
		updateData.Role, groupID, memberID)
	if err != nil {
		http.Error(w, "Failed to update member role", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Member not found or is the creator", http.StatusNotFound)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to complete role update", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member role updated successfully",
	})
}

func RemoveMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		sendJSONError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	memberID, err := strconv.Atoi(r.PathValue("memberId"))
	if err != nil {
		sendJSONError(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Remove member
	_, err = tx.Exec(`
		DELETE FROM group_members 
		WHERE group_id = ? AND user_id = ? AND role != 'creator'`,
		groupID, memberID)
	if err != nil {
		sendJSONError(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	// Clear any existing invitations or requests
	_, err = tx.Exec(`
		DELETE FROM group_invitations 
		WHERE group_id = ? AND invitee_id = ?`,
		groupID, memberID)
	if err != nil {
		log.Printf("Failed to clear invitations: %v", err)
	}

	// Create notification for removed member
	_, err = tx.Exec(`
		INSERT INTO notifications (
			user_id, 
			type, 
			content, 
			group_id, 
			created_at
		) VALUES (
			?, 
			'group_removal', 
			'You have been removed from the group. You can request to join again or wait for a new invitation.', 
			?, 
			CURRENT_TIMESTAMP
		)`,
		memberID, groupID)
	if err != nil {
		log.Printf("Failed to create notification: %v", err)
	}

	if err = tx.Commit(); err != nil {
		sendJSONError(w, "Failed to complete member removal", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Member removed successfully",
	})
}

func GetGroupRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		sendJSONError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get current user's role
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Check if user has permission to view requests
	hasPermission, err := checkUserRole(groupID, userID, "admin")
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !hasPermission {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get pending requests with user information
	rows, err := sqlite.DB.Query(`
		SELECT 
			gi.id,
			u.username,
			gi.created_at
		FROM group_invitations gi
		JOIN users u ON u.id = gi.invitee_id
		WHERE gi.group_id = ? 
		AND gi.type = 'request'
		AND gi.status = 'pending'
		ORDER BY gi.created_at DESC`,
		groupID)
	if err != nil {
		sendJSONError(w, "Failed to fetch requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []map[string]interface{}
	for rows.Next() {
		var request struct {
			ID        int
			Username  string
			CreatedAt time.Time
		}
		if err := rows.Scan(&request.ID, &request.Username, &request.CreatedAt); err != nil {
			log.Printf("Error scanning request: %v", err)
			continue
		}

		requests = append(requests, map[string]interface{}{
			"id":         request.ID,
			"username":   request.Username,
			"created_at": request.CreatedAt,
		})
	}

	if requests == nil {
		requests = make([]map[string]interface{}, 0)
	}

	sendJSONResponse(w, http.StatusOK, requests)
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
		sendJSONError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check if user is already a member
	var isMember bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_members 
			WHERE group_id = ? AND user_id = ?
		)`, groupID, userID).Scan(&isMember)
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if isMember {
		sendJSONError(w, "User is already a member", http.StatusBadRequest)
		return
	}

	// Check for existing request
	var hasRequest bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_invitations 
			WHERE group_id = ? AND invitee_id = ? AND type = 'request' AND status = 'pending'
		)`, groupID, userID).Scan(&hasRequest)
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	if hasRequest {
		sendJSONError(w, "User already has a pending request", http.StatusBadRequest)
		return
	}

	// Get group information
	var group struct {
		Title     string
		CreatorID int
	}
	err = tx.QueryRow(`SELECT title, creator_id FROM groups WHERE id = ?`, groupID).Scan(&group.Title, &group.CreatorID)
	if err != nil {
		sendJSONError(w, "Failed to get group information", http.StatusInternalServerError)
		return
	}

	// Create join request
	result, err := tx.Exec(`
		INSERT INTO group_invitations (
			group_id, 
			inviter_id, 
			invitee_id, 
			type,
			status, 
			created_at
		) VALUES (?, ?, ?, 'request', 'pending', CURRENT_TIMESTAMP)`,
		groupID, userID, userID)
	if err != nil {
		sendJSONError(w, "Failed to create join request", http.StatusInternalServerError)
		return
	}

	requestID, err := result.LastInsertId()
	if err != nil {
		sendJSONError(w, "Failed to get request ID", http.StatusInternalServerError)
		return
	}

	// Create notification for group creator
	_, err = tx.Exec(`
		INSERT INTO notifications (
			user_id,
			type,
			content,
			group_id,
			from_user_id,
			is_read,
			created_at
		) VALUES (?, ?, ?, ?, ?, false, CURRENT_TIMESTAMP)`,
		group.CreatorID,
		"group_join_request",
		fmt.Sprintf("%s has requested to join %s", username, group.Title),
		groupID,
		userID)
	if err != nil {
		log.Printf("Failed to create notification: %v", err)
	}

	if err = tx.Commit(); err != nil {
		sendJSONError(w, "Failed to complete join request", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":   "Join request sent successfully",
		"requestId": requestID,
	})
}

// HandleInvitation handles accepting or rejecting an invitation
func HandleInvitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get parameters from URL
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Printf("Invalid group ID: %v, raw value: %s", err, r.PathValue("id"))
		sendJSONError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	invitationID, err := strconv.Atoi(r.PathValue("invitationId"))
	if err != nil {
		log.Printf("Invalid invitation ID: %v, raw value: %s", err, r.PathValue("invitationId"))
		sendJSONError(w, "Invalid invitation ID", http.StatusBadRequest)
		return
	}

	action := r.PathValue("action")
	if action != "accept" && action != "reject" {
		log.Printf("Invalid action: %s", action)
		sendJSONError(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Debug log all request details
	log.Printf("Processing invitation request: method=%s, path=%s, groupID=%d, invitationID=%d, action=%s",
		r.Method, r.URL.Path, groupID, invitationID, action)

	// Verify the invitation exists first
	var exists bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 
			FROM group_invitations 
			WHERE id = ? AND group_id = ?
		)`, invitationID, groupID).Scan(&exists)

	if err != nil {
		log.Printf("Error checking invitation existence: %v", err)
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		log.Printf("Invitation not found: id=%d, groupID=%d", invitationID, groupID)
		sendJSONError(w, "Invitation not found", http.StatusNotFound)
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("Unauthorized: %v", err)
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Verify invitation exists and belongs to this user
	var invitation struct {
		Status string
		Type   string
	}
	err = tx.QueryRow(`
		SELECT status, type 
		FROM group_invitations 
		WHERE id = ? AND group_id = ? AND invitee_id = ? AND status = 'pending'`,
		invitationID, groupID, userID).Scan(&invitation.Status, &invitation.Type)

	// Debug log the invitation check
	log.Printf("Checking invitation: id=%d, groupID=%d, userID=%d, err=%v",
		invitationID, groupID, userID, err)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Invitation not found or already processed: %v", err)
			sendJSONError(w, "Invitation not found or already processed", http.StatusNotFound)
		} else {
			log.Printf("Database error checking invitation: %v", err)
			sendJSONError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if action == "accept" {
		// Add user as group member
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, user_id, role, joined_at)
			VALUES (?, ?, 'member', CURRENT_TIMESTAMP)`,
			groupID, userID)
		if err != nil {
			sendJSONError(w, "Failed to add member to group", http.StatusInternalServerError)
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
		sendJSONError(w, "Failed to update invitation status", http.StatusInternalServerError)
		return
	}

	// Delete the notification
	_, err = tx.Exec(`
		DELETE FROM notifications 
		WHERE type = 'group_invitation' 
		AND group_id = ? 
		AND invitation_id = ?`,
		groupID, invitationID)
	if err != nil {
		log.Printf("Failed to delete notification: %v", err)
	}

	// Get the invitation details including inviter_id and group title
	var inviterID int
	var groupTitle string
	err = tx.QueryRow(`
		SELECT i.inviter_id, g.title 
		FROM group_invitations i 
		JOIN groups g ON i.group_id = g.id 
		WHERE i.id = ?`, invitationID).Scan(&inviterID, &groupTitle)
	if err != nil {
		log.Printf("Failed to get invitation details: %v", err)
		sendJSONError(w, "Failed to get invitation details", http.StatusInternalServerError)
		return
	}

	// Create a notification for the inviter about the action
	_, err = tx.Exec(`
		INSERT INTO notifications (
			user_id,
			type,
			content,
			group_id,
			from_user_id,
			created_at
		) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		inviterID,
		"invitation_response",
		fmt.Sprintf("%s has %sed your invitation to join %s", username, action, groupTitle),
		groupID,
		inviterID)
	if err != nil {
		log.Printf("Failed to create response notification: %v", err)
	}

	if err = tx.Commit(); err != nil {
		sendJSONError(w, "Failed to complete action", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Invitation %sed successfully", action),
	})
}

func CreatePostComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.PathValue("postId"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
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

	// Verify user is a member
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

	// Parse request body
	var comment struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create comment
	result, err := sqlite.DB.Exec(`
		INSERT INTO group_post_comments (post_id, author_id, content)
		VALUES (?, ?, ?)`,
		postID, userID, comment.Content)
	if err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	commentID, _ := result.LastInsertId()

	// Get the created comment with author info
	var createdComment struct {
		ID        int64  `json:"id"`
		Content   string `json:"content"`
		AuthorID  int64  `json:"author_id"`
		Author    string `json:"author"`
		CreatedAt string `json:"created_at"`
	}

	err = sqlite.DB.QueryRow(`
		SELECT c.id, c.content, c.author_id, u.username, c.created_at
		FROM group_post_comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.id = ?`,
		commentID).Scan(&createdComment.ID, &createdComment.Content, &createdComment.AuthorID, &createdComment.Author, &createdComment.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to fetch created comment", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(createdComment)
}

func GetGroupPostComments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.PathValue("postId"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
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

	// Verify user is a member
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

	// Get comments
	rows, err := sqlite.DB.Query(`
		SELECT c.id, c.content, c.author_id, u.username, c.created_at
		FROM group_post_comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC`,
		postID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []map[string]interface{}
	for rows.Next() {
		var comment struct {
			ID        int64  `json:"id"`
			Content   string `json:"content"`
			AuthorID  int64  `json:"author_id"`
			Author    string `json:"author"`
			CreatedAt string `json:"created_at"`
		}
		err := rows.Scan(&comment.ID, &comment.Content, &comment.AuthorID, &comment.Author, &comment.CreatedAt)
		if err != nil {
			continue
		}
		comments = append(comments, map[string]interface{}{
			"id":         comment.ID,
			"content":    comment.Content,
			"author_id":  comment.AuthorID,
			"author":     comment.Author,
			"created_at": comment.CreatedAt,
		})
	}

	json.NewEncoder(w).Encode(comments)
}

func HandleJoinRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get group ID from URL
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get action from URL
	action := r.PathValue("action")
	if action != "accept" && action != "reject" {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request struct {
		RequestID int `json:"requestId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
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

	// Verify the request belongs to this group
	var joinRequest struct {
		ID        int
		UserID    int
		GroupID   int
		Status    string
		CreatedAt string
	}
	err = tx.QueryRow(`
		SELECT id, invitee_id, group_id, status, created_at 
		FROM group_invitations 
		WHERE id = ? AND group_id = ? AND status = 'pending'`,
		request.RequestID, groupID).Scan(
		&joinRequest.ID,
		&joinRequest.UserID, // invitee_id maps to UserID
		&joinRequest.GroupID,
		&joinRequest.Status,
		&joinRequest.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Join request not found or already processed", http.StatusNotFound)
		} else {
			log.Printf("Database error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if action == "accept" {
		// Check if user is already a member
		var isMember bool
		err = tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM group_members 
				WHERE group_id = ? AND user_id = ?
			)`, joinRequest.GroupID, joinRequest.UserID).Scan(&isMember)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if isMember {
			http.Error(w, "User is already a member", http.StatusBadRequest)
			return
		}

		// Add user as group member
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, user_id, role, joined_at)
			VALUES (?, ?, 'member', CURRENT_TIMESTAMP)`,
			joinRequest.GroupID, joinRequest.UserID)
		if err != nil {
			http.Error(w, "Failed to add member to group", http.StatusInternalServerError)
			return
		}
	}

	// Update request status
	_, err = tx.Exec(`
		UPDATE group_invitations 
		SET status = ? 
		WHERE id = ?`,
		action+"ed", request.RequestID)
	if err != nil {
		http.Error(w, "Failed to update request status", http.StatusInternalServerError)
		return
	}

	// Delete related notifications
	_, err = tx.Exec(`
		DELETE FROM notifications 
		WHERE type = 'group_join_request' 
		AND group_id = ? 
		AND user_id = ?`,
		joinRequest.GroupID, joinRequest.UserID)
	if err != nil {
		log.Printf("Failed to delete notifications: %v", err)
	}

	// Create notification for the requester
	if action == "accept" {
		_, err = tx.Exec(`
			INSERT INTO notifications (user_id, type, content, group_id, created_at)
			VALUES (?, 'group_join_accepted', 'Your request to join the group has been accepted', ?, CURRENT_TIMESTAMP)`,
			joinRequest.UserID, joinRequest.GroupID)
		if err != nil {
			log.Printf("Failed to create acceptance notification: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to complete action", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Request %sed successfully", action),
	})
}

func GetMemberRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		sendJSONError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Get user's role
	var role string
	err = sqlite.DB.QueryRow(`
		SELECT role FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJSONError(w, "Not a member of this group", http.StatusNotFound)
		} else {
			sendJSONError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{
		"role": role,
	})
}
