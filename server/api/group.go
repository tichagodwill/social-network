package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"social-network/models"
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

	// Step 1: Create a chat for the group
	result, err := tx.Exec(`
       INSERT INTO chats (type, created_at)
       VALUES ('group', CURRENT_TIMESTAMP)`)
	if err != nil {
		http.Error(w, "Failed to create group chat", http.StatusInternalServerError)
		return
	}

	// Get the new chat ID
	chatID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get new chat ID", http.StatusInternalServerError)
		return
	}

	// Step 2: Create the group with a reference to the chat
	result, err = tx.Exec(`
       INSERT INTO groups (title, description, creator_id, chat_id)
       VALUES (?, ?, ?, ?)`,
		group.Title, group.Description, creatorID, chatID)
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

	// Step 3: Add creator as a member with creator role
	_, err = tx.Exec(`
       INSERT INTO group_members (group_id, user_id, role)
       VALUES (?, ?, 'creator')`,
		groupID, creatorID)
	if err != nil {
		http.Error(w, "Failed to add creator as a member", http.StatusInternalServerError)
		return
	}

	// Step 4: Add creator to the group chat
	_, err = tx.Exec(`
       INSERT INTO user_chat_status (user_id, chat_id)
       VALUES (?, ?)`,
		creatorID, chatID)
	if err != nil {
		http.Error(w, "Failed to add creator to chat", http.StatusInternalServerError)
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
		"chatId":  chatID,
	})
}

func CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with 10MB limit
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, fmt.Sprintf("Error parsing form: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Form values received: %+v", r.Form)
	log.Printf("Files received: %+v", r.MultipartForm.File)

	// Get group ID and validate it
	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		log.Printf("Invalid group ID: %v", err)
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("Session error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var authorID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&authorID)
	if err != nil {
		log.Printf("Error getting author ID: %v", err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Check if user is a member of the group
	var isMember bool
	err = sqlite.DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM group_members 
            WHERE group_id = ? AND user_id = ?
        )`, groupID, authorID).Scan(&isMember)
	if err != nil {
		log.Printf("Error checking membership: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		log.Printf("User %d is not a member of group %d", authorID, groupID)
		http.Error(w, "Not a group member", http.StatusForbidden)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	log.Printf("Creating post - Title: %s, Content: %s", title, content)

	var mediaPath string
	// Handle file upload if present
	file, header, err := r.FormFile("media")
	if err != nil {
		if err != http.ErrMissingFile {
			log.Printf("Error getting file from form: %v", err)
			http.Error(w, fmt.Sprintf("Error processing file: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		defer file.Close()
		log.Printf("Received file: %s, size: %d, type: %s",
			header.Filename,
			header.Size,
			header.Header.Get("Content-Type"))

		// Validate file type
		fileType := header.Header.Get("Content-Type")
		allowedTypes := map[string]bool{
			"image/jpeg":      true,
			"image/png":       true,
			"image/gif":       true,
			"application/pdf": true,
		}

		if !allowedTypes[fileType] {
			log.Printf("Invalid file type: %s", fileType)
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		// Create uploads directory if it doesn't exist
		uploadDir := "./uploads/group_posts"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Printf("Error creating upload directory: %v", err)
			http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
			return
		}

		// Generate unique filename
		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("%d_%d_%s%s", groupID, time.Now().Unix(), util.GenerateRandomString(8), ext)
		fullPath := filepath.Join(uploadDir, filename)
		log.Printf("Saving file to: %s", fullPath)

		dst, err := os.Create(fullPath)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			http.Error(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		written, err := io.Copy(dst, file)
		if err != nil {
			log.Printf("Error copying file: %v", err)
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		log.Printf("Successfully wrote %d bytes to %s", written, fullPath)

		mediaPath = filename
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create post
	result, err := tx.Exec(`
        INSERT INTO group_posts (group_id, author_id, title, content, media, created_at)
        VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		groupID, authorID, title, content, mediaPath)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	postID, _ := result.LastInsertId()
	log.Printf("Post created with ID: %d", postID)

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to complete post creation", http.StatusInternalServerError)
		return
	}

	// Return the created post
	var post struct {
		ID        int64  `json:"id"`
		GroupID   int    `json:"group_id"`
		AuthorID  int    `json:"author_id"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		Media     string `json:"media,omitempty"`
		CreatedAt string `json:"created_at"`
	}

	err = sqlite.DB.QueryRow(`
        SELECT id, group_id, author_id, title, content, media, created_at
        FROM group_posts
        WHERE id = ?
    `, postID).Scan(&post.ID, &post.GroupID, &post.AuthorID, &post.Title, &post.Content, &post.Media, &post.CreatedAt)

	if err != nil {
		log.Printf("Error fetching created post: %v", err)
		http.Error(w, "Failed to fetch created post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
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
		SELECT p.id, p.group_id, p.author_id, u.username, p.title, p.content, p.media, p.created_at, p.updated_at
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
			&post.Title, &post.Content, &post.Media, &post.CreatedAt, &post.UpdatedAt)
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

	// Get path parameters
	groupID := r.PathValue("id")
	postID := r.PathValue("postId")
	log.Printf("Creating comment for group %s, post %s", groupID, postID)

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("Auth error: %v", err)
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		http.Error(w, `{"error": "Failed to get user information"}`, http.StatusInternalServerError)
		return
	}

	// Parse request body
	var commentData struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&commentData); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Comment data: %+v", commentData)

	// Validate content
	if strings.TrimSpace(commentData.Content) == "" {
		http.Error(w, `{"error": "Comment content is required"}`, http.StatusBadRequest)
		return
	}

	// Convert string IDs to integers
	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		log.Printf("Invalid group ID: %v", err)
		http.Error(w, `{"error": "Invalid group ID"}`, http.StatusBadRequest)
		return
	}

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		log.Printf("Invalid post ID: %v", err)
		http.Error(w, `{"error": "Invalid post ID"}`, http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// First verify the post exists
	log.Printf("Checking post existence for group %d, post %d", groupIDInt, postIDInt)
	var postExists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM group_posts 
			WHERE id = ? AND group_id = ?
		)`, postIDInt, groupIDInt).Scan(&postExists)
	if err != nil {
		log.Printf("Error checking post: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	if !postExists {
		log.Printf("Post not found: group %d, post %d", groupIDInt, postIDInt)
		http.Error(w, `{"error": "Post not found or doesn't belong to this group"}`, http.StatusNotFound)
		return
	}

	// Insert comment
	log.Printf("Inserting comment: post_id=%d, author_id=%d, content=%s",
		postIDInt, userID, commentData.Content)
	result, err := tx.Exec(`
		INSERT INTO group_post_comments (post_id, author_id, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		postIDInt, userID, commentData.Content)
	if err != nil {
		log.Printf("Error inserting comment: %v", err)
		http.Error(w, `{"error": "Failed to create comment"}`, http.StatusInternalServerError)
		return
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting comment ID: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	log.Printf("Created comment with ID: %d", commentID)

	// Get complete comment data
	var createdComment m.GroupPostComment
	err = tx.QueryRow(`
		SELECT 
			c.id,
			c.post_id,
			c.author_id,
			u.username as author,
			c.content,
			c.created_at,
			c.created_at as updated_at
		FROM group_post_comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.id = ?`,
		commentID).Scan(
		&createdComment.ID,
		&createdComment.PostID,
		&createdComment.AuthorID,
		&createdComment.Author,
		&createdComment.Content,
		&createdComment.CreatedAt,
		&createdComment.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error retrieving comment: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Return the created comment
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdComment)
}

func getPostComments(postID int) ([]m.GroupPostComment, error) {
	rows, err := sqlite.DB.Query(`
		SELECT 
			c.id,
			c.post_id,
			c.author_id,
			u.username as author,
			c.content,
			c.created_at,
			c.created_at as updated_at
		FROM group_post_comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC`,
		postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []m.GroupPostComment
	for rows.Next() {
		var comment m.GroupPostComment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&comment.Author,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning comment: %v", err)
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
		SendNotification([]int{inviteeID}, notification)
	}

	if err = tx.Commit(); err != nil {
		sendJSONError(w, "Failed to complete invitation", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Invitation sent successfully",
	})
}

func GroupAccept(w http.ResponseWriter, r *http.Request) {
	// Implementation for accepting group invitation
}

func GroupReject(w http.ResponseWriter, r *http.Request) {
	// Implementation for rejecting group invitation
}

func GroupLeave(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is the creator - creators cannot leave their own group
	var isCreator bool
	err = tx.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM group_members
            WHERE group_id = ? AND user_id = ? AND role = 'creator'
        )
    `, groupID, userID).Scan(&isCreator)

	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if isCreator {
		sendJSONError(w, "Group creators cannot leave their own group. Please delete the group or transfer ownership first.", http.StatusBadRequest)
		return
	}

	// Get the chat_id for this group
	var chatID int
	err = tx.QueryRow(`SELECT chat_id FROM groups WHERE id = ?`, groupID).Scan(&chatID)
	if err != nil {
		log.Printf("Failed to get chat ID for group %d: %v", groupID, err)
		sendJSONError(w, "Failed to get group chat information", http.StatusInternalServerError)
		return
	}

	// Remove member from group_members
	result, err := tx.Exec(`
        DELETE FROM group_members 
        WHERE group_id = ? AND user_id = ?`,
		groupID, userID)
	if err != nil {
		sendJSONError(w, "Failed to leave group", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendJSONError(w, "You are not a member of this group", http.StatusBadRequest)
		return
	}

	// CRITICAL: Remove member from user_chat_status to hide chat
	_, err = tx.Exec(`
        DELETE FROM user_chat_status
        WHERE chat_id = ? AND user_id = ?`,
		chatID, userID)
	if err != nil {
		log.Printf("Failed to remove user from user_chat_status: %v", err)
		// Continue even if this fails
	}

	// Get group info for notification to group admins
	var groupName string
	var creatorID int
	err = tx.QueryRow(`SELECT title, creator_id FROM groups WHERE id = ?`, groupID).Scan(&groupName, &creatorID)
	if err != nil {
		log.Printf("Failed to get group info: %v", err)
		groupName = "the group"
	}

	// Notify the group creator/admin about the user leaving
	_, err = tx.Exec(`
        INSERT INTO notifications (
            user_id, 
            type, 
            content, 
            group_id, 
            from_user_id,
            created_at
        ) VALUES (
            ?, 
            'group_member_left', 
            ?, 
            ?, 
            ?,
            CURRENT_TIMESTAMP
        )`,
		creatorID,
		fmt.Sprintf("%s has left %s", username, groupName),
		groupID,
		userID)
	if err != nil {
		log.Printf("Failed to create notification for group admin: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		sendJSONError(w, "Failed to complete leaving group", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("You have successfully left '%s'", groupName),
	})
}

func GetGroupEvents(w http.ResponseWriter, r *http.Request) {
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

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	rows, err := sqlite.DB.Query(`
		SELECT 
			e.id,
			e.title,
			e.description,
			e.event_date,
			e.creator_id,
			COALESCE(ur.rsvp_status, '') as user_response,
			COALESCE(r_going.going_count, 0) as going_count,
			COALESCE(r_not_going.not_going_count, 0) as not_going_count
		FROM group_events e
		LEFT JOIN group_event_RSVP ur ON e.id = ur.event_id AND ur.user_id = ?
		LEFT JOIN (
			SELECT event_id, COUNT(*) as going_count 
			FROM group_event_RSVP 
			WHERE rsvp_status = 'going' 
			GROUP BY event_id
		) r_going ON e.id = r_going.event_id
		LEFT JOIN (
			SELECT event_id, COUNT(*) as not_going_count 
			FROM group_event_RSVP 
			WHERE rsvp_status = 'not_going' 
			GROUP BY event_id
		) r_not_going ON e.id = r_not_going.event_id
		WHERE e.group_id = ?
		ORDER BY e.event_date DESC`,
		userID, groupID)
	if err != nil {
		log.Printf("Error querying events: %v", err)
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var (
			id        int
			title     string
			desc      string
			date      time.Time
			creatorID int
			userResp  string
			going     int
			notGoing  int
		)

		if err := rows.Scan(&id, &title, &desc, &date, &creatorID, &userResp, &going, &notGoing); err != nil {
			log.Printf("Error scanning event: %v", err)
			continue
		}

		event := map[string]interface{}{
			"id":           id,
			"title":        title,
			"description":  desc,
			"eventDate":    date.Format(time.RFC3339), // Format the date as ISO string
			"creatorId":    creatorID,
			"userResponse": userResp,
			"responses": map[string]int{
				"going":    going,
				"notGoing": notGoing,
			},
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating events: %v", err)
		http.Error(w, "Error processing events", http.StatusInternalServerError)
		return
	}

	// If no events found, return empty array instead of null
	if events == nil {
		events = make([]map[string]interface{}, 0)
	}

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

	// After successfully creating the event, notify group members
	go func() {
		log.Printf("Starting notification process for event '%s' in group %d", event.Title, groupID)

		// Get all group members immediately
		rows, err := sqlite.DB.Query(`
			SELECT user_id 
			FROM group_members 
			WHERE group_id = ? AND user_id != ?`,
			groupID, event.CreatorID)
		if err != nil {
			log.Printf("Error getting group members: %v", err)
			return
		}
		defer rows.Close()

		var groupName string
		err = sqlite.DB.QueryRow("SELECT title FROM groups WHERE id = ?", groupID).Scan(&groupName)
		if err != nil {
			log.Printf("Error getting group name: %v", err)
			return
		}

		// Create notifications in a transaction
		tx, err := sqlite.DB.Begin()
		if err != nil {
			log.Printf("Error starting transaction: %v", err)
			return
		}

		for rows.Next() {
			var memberID int
			if err := rows.Scan(&memberID); err != nil {
				log.Printf("Error scanning member ID: %v", err)
				continue
			}

			// Insert notification
			result, err := tx.Exec(`
				INSERT INTO notifications (
					user_id, type, content, group_id, created_at, is_read
				) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, false)`,
				memberID,
				"group_event",
				fmt.Sprintf("New event '%s' created in group '%s'", event.Title, groupName),
				groupID,
			)
			if err != nil {
				log.Printf("Error creating notification for member %d: %v", memberID, err)
				continue
			}

			notificationID, _ := result.LastInsertId()

			// Broadcast immediately
			notification := models.WebSocketMessage{
				Type: "notification",
				Data: map[string]interface{}{
					"id":        notificationID,
					"type":      "group_event",
					"content":   fmt.Sprintf("New event '%s' created in group '%s'", event.Title, groupName),
					"groupId":   groupID,
					"link":      fmt.Sprintf("/groups/%d", groupID),
					"isRead":    false,
					"createdAt": time.Now().Format(time.RFC3339),
					"userId":    memberID, // Add this to ensure proper routing
				},
			}

			log.Printf("Sending event notification to member %d: %+v", memberID, notification)
			broadcast <- models.BroadcastMessage{
				Data:        notification,
				TargetUsers: map[int]bool{memberID: true},
			}
		}

		if err = tx.Commit(); err != nil {
			log.Printf("Error committing notifications: %v", err)
			tx.Rollback()
			return
		}
	}()

	log.Printf("Event created successfully, notification process started in background")
	// Return the created event
	json.NewEncoder(w).Encode(event)
}

// Update the RespondToGroupEvent function to use these helpers
func RespondToGroupEvent(w http.ResponseWriter, r *http.Request) {
	log.Println("1. Starting RespondToGroupEvent handler")

	// Get parameters from URL
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Printf("2. Invalid group ID: %v", err)
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}
	log.Printf("3. Group ID: %d", groupID)

	eventID, err := strconv.Atoi(r.PathValue("eventId"))
	if err != nil {
		log.Printf("4. Invalid event ID: %v", err)
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}
	log.Printf("5. Event ID: %d", eventID)

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("6. Auth error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("7. Username: %s", username)

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("8. User lookup error: %v", err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}
	log.Printf("9. User ID: %d", userID)

	// Parse request body
	var requestBody struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("10. Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("11. Request status: %s", requestBody.Status)

	// Validate status
	if requestBody.Status != "going" && requestBody.Status != "not_going" {
		log.Printf("12. Invalid status value: %s", requestBody.Status)
		http.Error(w, "Invalid status. Must be 'going' or 'not_going'", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("13. Transaction start error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Verify event exists
	var eventExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM group_events WHERE id = ? AND group_id = ?)",
		eventID, groupID).Scan(&eventExists)
	if err != nil {
		log.Printf("14. Error checking event: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !eventExists {
		log.Printf("15. Event not found")
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Update or insert RSVP
	_, err = tx.Exec(`
		INSERT INTO group_event_RSVP (event_id, user_id, rsvp_status)
		VALUES (?, ?, ?)
		ON CONFLICT(event_id, user_id) 
		DO UPDATE SET rsvp_status = ?`,
		eventID, userID, requestBody.Status, requestBody.Status)
	if err != nil {
		log.Printf("16. RSVP update error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to update RSVP",
		})
		return
	}

	// Get updated counts
	var going, notGoing int
	err = tx.QueryRow(`
		SELECT 
			COUNT(CASE WHEN rsvp_status = 'going' THEN 1 END) as going,
			COUNT(CASE WHEN rsvp_status = 'not_going' THEN 1 END) as not_going
		FROM group_event_RSVP 
		WHERE event_id = ?`,
		eventID).Scan(&going, &notGoing)
	if err != nil {
		log.Printf("17. Count query error: %v", err)
		http.Error(w, "Failed to get updated counts", http.StatusInternalServerError)
		return
	}
	log.Printf("18. Counts - Going: %d, Not Going: %d", going, notGoing)

	if err = tx.Commit(); err != nil {
		log.Printf("19. Transaction commit error: %v", err)
		http.Error(w, "Failed to complete action", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message":  "RSVP updated successfully",
		"going":    going,
		"notGoing": notGoing,
	}
	log.Printf("20. Sending response: %+v", response)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("21. Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("22. Handler completed successfully")
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

	// Get the chat ID associated with this group
	var chatID int
	err = sqlite.DB.QueryRow("SELECT chat_id FROM groups WHERE id = ?", groupID).Scan(&chatID)
	if err != nil {
		http.Error(w, "Group not found or has no associated chat", http.StatusNotFound)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete all related data
	// The order matters here due to foreign key constraints

	// Delete group members
	_, err = tx.Exec("DELETE FROM group_members WHERE group_id = ?", groupID)
	if err != nil {
		http.Error(w, "Failed to delete group members", http.StatusInternalServerError)
		return
	}

	// Delete events RSVPs
	_, err = tx.Exec(`
        DELETE FROM group_event_RSVP 
        WHERE event_id IN (SELECT id FROM group_events WHERE group_id = ?)`,
		groupID)
	if err != nil {
		http.Error(w, "Failed to delete event RSVPs", http.StatusInternalServerError)
		return
	}

	// Delete events
	_, err = tx.Exec("DELETE FROM group_events WHERE group_id = ?", groupID)
	if err != nil {
		http.Error(w, "Failed to delete events", http.StatusInternalServerError)
		return
	}

	// Delete the group
	_, err = tx.Exec("DELETE FROM groups WHERE id = ?", groupID)
	if err != nil {
		http.Error(w, "Failed to delete group", http.StatusInternalServerError)
		return
	}

	// Delete user chat statuses
	_, err = tx.Exec("DELETE FROM user_chat_status WHERE chat_id = ?", chatID)
	if err != nil {
		http.Error(w, "Failed to delete chat participants", http.StatusInternalServerError)
		return
	}

	// Delete chat messages
	_, err = tx.Exec("DELETE FROM chat_messages WHERE chat_id = ?", chatID)
	if err != nil {
		http.Error(w, "Failed to delete chat messages", http.StatusInternalServerError)
		return
	}

	// Finally delete the chat itself
	_, err = tx.Exec("DELETE FROM chats WHERE id = ?", chatID)
	if err != nil {
		http.Error(w, "Failed to delete chat", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to complete deletion", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Group and associated chat deleted successfully",
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

	// Get the chat_id for this group
	var chatID int
	err = tx.QueryRow(`SELECT chat_id FROM groups WHERE id = ?`, groupID).Scan(&chatID)
	if err != nil {
		log.Printf("Failed to get chat ID for group %d: %v", groupID, err)
		sendJSONError(w, "Failed to get group chat information", http.StatusInternalServerError)
		return
	}

	// Remove member from group_members
	_, err = tx.Exec(`
        DELETE FROM group_members 
        WHERE group_id = ? AND user_id = ? AND role != 'creator'`,
		groupID, memberID)
	if err != nil {
		sendJSONError(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	// CRITICAL: Remove member from user_chat_status to hide chat
	_, err = tx.Exec(`
        DELETE FROM user_chat_status
        WHERE chat_id = ? AND user_id = ?`,
		chatID, memberID)
	if err != nil {
		log.Printf("Failed to remove user from user_chat_status: %v", err)
		// Continue even if this fails - it's better to have the user still see the chat
		// than to erroneously keep them as a group member
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

	// Fix syntax errors in notification creation
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
		// Get the chat_id for this group
		var chatID int
		err = tx.QueryRow(`SELECT chat_id FROM groups WHERE id = ?`, groupID).Scan(&chatID)
		if err != nil {
			log.Printf("Failed to get chat ID for group %d: %v", groupID, err)
			sendJSONError(w, "Failed to get group chat information", http.StatusInternalServerError)
			return
		}

		// Add user as group member
		_, err = tx.Exec(`
            INSERT INTO group_members (group_id, user_id, role, joined_at)
            VALUES (?, ?, 'member', CURRENT_TIMESTAMP)`,
			groupID, userID)
		if err != nil {
			log.Printf("Failed to add user to group_members: %v", err)
			sendJSONError(w, "Failed to add member to group", http.StatusInternalServerError)
			return
		}

		// CRITICAL: Add user to user_chat_status to ensure chat visibility
		_, err = tx.Exec(`
            INSERT INTO user_chat_status (user_id, chat_id)
            VALUES (?, ?)
            ON CONFLICT(user_id, chat_id) DO NOTHING`,
			userID, chatID)
		if err != nil {
			log.Printf("Failed to add user to user_chat_status: %v", err)
			sendJSONError(w, "Failed to add member to group chat", http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully added user %d to group %d with chat_id %d", userID, groupID, chatID)
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

		// Get the chat_id for this group
		var chatID int
		err = tx.QueryRow(`SELECT chat_id FROM groups WHERE id = ?`, joinRequest.GroupID).Scan(&chatID)
		if err != nil {
			log.Printf("Failed to get chat ID for group %d: %v", joinRequest.GroupID, err)
			http.Error(w, "Failed to get group chat information", http.StatusInternalServerError)
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

		// CRITICAL: Add user to user_chat_status to ensure chat visibility
		_, err = tx.Exec(`
            INSERT INTO user_chat_status (user_id, chat_id)
            VALUES (?, ?)
            ON CONFLICT(user_id, chat_id) DO NOTHING`,
			joinRequest.UserID, chatID)
		if err != nil {
			log.Printf("Failed to add user to user_chat_status: %v", err)
			http.Error(w, "Failed to add member to group chat", http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully added user %d to group %d with chat_id %d",
			joinRequest.UserID, joinRequest.GroupID, chatID)
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

	groupID := r.PathValue("id")
	if groupID == "" {
		sendJSONError(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		sendJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	var role string
	err = sqlite.DB.QueryRow(`
		SELECT role FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		groupID, userID).Scan(&role)

	if err == sql.ErrNoRows {
		sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"role": nil,
		})
		return
	}

	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"role": role,
	})
}

func ServeGroupPostMedia(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "No filename provided", http.StatusBadRequest)
		return
	}

	// Sanitize filename to prevent directory traversal
	filename = filepath.Base(filename)

	filePath := filepath.Join("./uploads/group_posts", filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}
