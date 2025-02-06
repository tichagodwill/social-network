package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
)

func CreateComment(w http.ResponseWriter, r *http.Request) {
	log.Println("=== CREATE COMMENT HANDLER CALLED ===")
	w.Header().Set("Content-Type", "application/json")

	// Get and log path parameters
	groupID := r.PathValue("id")
	postID := r.PathValue("postId")
	log.Printf("Creating comment for group %s, post %s", groupID, postID)

	// Log request headers and body
	log.Printf("Request headers: %+v", r.Header)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	log.Printf("Request body: %s", string(body))
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset the body for later use

	var comment m.Comment
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&comment); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed comment: %+v", comment)

	// Validate required fields
	if comment.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}
	if comment.PostID == 0 {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	if comment.Author == 0 {
		http.Error(w, "Author ID is required", http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Verify the post exists
	var postExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", comment.PostID).Scan(&postExists)
	if err != nil {
		log.Printf("Error checking post existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !postExists {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Insert the comment
	result, err := tx.Exec(`
		INSERT INTO comments (content, media, author, post_id, created_at) 
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		comment.Content, comment.Media, comment.Author, comment.PostID)
	if err != nil {
		log.Printf("Error inserting comment: %v", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Get the inserted ID
	commentID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("Comment inserted with ID: %d", commentID)

	// Verify the comment was saved
	var savedComment m.Comment
	err = tx.QueryRow(`
		SELECT id, content, post_id, author 
		FROM comments 
		WHERE id = ?`, commentID).Scan(
		&savedComment.ID,
		&savedComment.Content,
		&savedComment.PostID,
		&savedComment.Author,
	)
	if err != nil {
		log.Printf("Error verifying saved comment: %v", err)
	} else {
		log.Printf("Verified saved comment: %+v", savedComment)
	}

	// Get the complete comment data
	var createdComment m.Comment
	err = tx.QueryRow(`
		SELECT 
			c.id,
			c.content,
			c.media,
			c.post_id,
			c.author,
			c.created_at,
			u.username as author_name,
			u.avatar as author_avatar
		FROM comments c
		JOIN users u ON c.author = u.id
		WHERE c.id = ?`,
		commentID).Scan(
		&createdComment.ID,
		&createdComment.Content,
		&createdComment.Media,
		&createdComment.PostID,
		&createdComment.Author,
		&createdComment.CreatedAt,
		&createdComment.AuthorName,
		&createdComment.AuthorAvatar,
	)
	if err != nil {
		log.Printf("Error retrieving created comment: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return the created comment
	json.NewEncoder(w).Encode(createdComment)
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	log.Println("=== GET COMMENTS HANDLER CALLED ===")
	w.Header().Set("Content-Type", "application/json")

	var comments []m.Comment
	postIDString := r.PathValue("postID")

	postID, err := strconv.Atoi(postIDString)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	rows, err := sqlite.DB.Query(`
		SELECT 
			c.id,
			c.content,
			c.media,
			c.post_id,
			c.author,
			c.created_at,
			u.username as author_name,
			u.avatar as author_avatar
		FROM comments c
		JOIN users u ON c.author = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC`,
		postID)
	if err != nil {
		if err == sql.ErrNoRows {
			json.NewEncoder(w).Encode([]m.Comment{})
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var comment m.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.Media,
			&comment.PostID,
			&comment.Author,
			&comment.CreatedAt,
			&comment.AuthorName,
			&comment.AuthorAvatar,
		)
		if err != nil {
			http.Error(w, "Error reading comment data", http.StatusInternalServerError)
			return
		}
		comments = append(comments, comment)
	}

	json.NewEncoder(w).Encode(comments)
}
