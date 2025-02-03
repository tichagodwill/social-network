package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

// getMyPosts fetches all posts created by the current user
func GetMyPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Fetch posts from the database
	rows, err := sqlite.DB.Query(`
		SELECT p.id, p.title, p.content, p.media, p.privacy, p.author, p.created_at, p.group_id
		FROM posts p
		WHERE p.author = ?
		ORDER BY p.created_at DESC`,
		userID)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		log.Printf("Error fetching posts: %v", err)
		return
	}

	defer rows.Close()

	// Iterate through the results
	var posts []m.Post
	for rows.Next() {
		var post m.Post
		var groupID *int // Use a pointer for the nullable GroupID

		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privacy, &post.Author, &post.CreatedAt, &groupID); err != nil {
			http.Error(w, "Error reading posts", http.StatusInternalServerError)
			log.Printf("Error scanning post: %v", err)
			return
		}

		// Fetch the author's username from the database
		var authorName string
		var authorAvatar string
		err = sqlite.DB.QueryRow(`
		SELECT username, avatar
		FROM users
		WHERE id = ?`,
			post.Author).Scan(&authorName, &authorAvatar)
		if err != nil {
			http.Error(w, "Failed to fetch author's username", http.StatusInternalServerError)
			log.Printf("Error fetching author's username: %v", err)
			return
		}

		post.AuthorName = authorName
		post.AuthorAvatar = authorAvatar

		// Now set the GroupID properly (can be nil if the database value is NULL)
		if groupID != nil {
			post.GroupID = *groupID
		} else {
			post.GroupID = 0 // Or set it to a default value if appropriate
		}

		posts = append(posts, post)
	}

	// Return the posts as JSON
	json.NewEncoder(w).Encode(posts)

}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user information",
		})
		return
	}

	var post m.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON data",
		})
		return
	}

	post.Author = userID
	// Validate required fields
	if strings.TrimSpace(post.Title) == "" || strings.TrimSpace(post.Content) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Title and content cannot be empty",
		})
		return
	}

	// Set default privacy if not provided
	if post.Privacy == 0 {
		post.Privacy = 1 // 1 for public, 2 for private, 3 for followers only
	}

	// Start a transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback() // Rollback if we don't commit

	// Insert the post into the database using the transaction
	result, err := tx.Exec(
		"INSERT INTO posts (title, content, media, privacy, author, created_at) VALUES (?, ?, ?, ?, ?, datetime('now'))",
		post.Title, post.Content, post.Media, post.Privacy, post.Author)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create post",
		})
		log.Printf("Error creating post: %v", err)
		return
	}

	// Get the ID of the newly created post
	postID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get post ID",
		})
		return
	}

	// Handle private post viewers
	if post.Privacy == 3 && len(post.SelectedUsers) > 0 {
		// Insert into post_PrivateViews for each selected user
		for _, selectedUserID := range post.SelectedUsers {
			_, err = tx.Exec(
				"INSERT INTO post_PrivateViews (post_id, user_id) VALUES (?, ?)",
				postID, selectedUserID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Failed to set post privacy views",
				})
				return
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to commit transaction",
		})
		return
	}

	// After successfully creating the post, fetch the complete post data
	var completePost m.Post
	err = sqlite.DB.QueryRow(`
        SELECT p.id, p.title, p.content, p.media, p.privacy, p.author, p.created_at,
               u.username as authorName, u.avatar as authorAvatar
        FROM posts p
        JOIN users u ON p.author = u.id
        WHERE p.id = ?`,
		postID).Scan(
		&completePost.ID,
		&completePost.Title,
		&completePost.Content,
		&completePost.Media,
		&completePost.Privacy,
		&completePost.Author,
		&completePost.CreatedAt,
		&completePost.AuthorName,
		&completePost.AuthorAvatar,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch created post",
		})
		return
	}

	// Return the complete post data
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(completePost)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user information",
		})
		return
	}

	// Check the database connection
	if sqlite.DB == nil {
		log.Printf("Error: Database is not initialized")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Database not initialized",
		})
		return
	}

	// Updated query to handle all privacy cases
	rows, err := sqlite.DB.Query(`
        SELECT DISTINCT p.id, p.title, p.content, p.media, p.privacy, p.author, p.created_at, p.group_id
        FROM posts p
        LEFT JOIN followers f ON p.author = f.followed_id AND f.follower_id = ?
        LEFT JOIN post_PrivateViews pv ON p.id = pv.post_id AND pv.user_id = ?
        WHERE p.privacy = 1  -- Public posts
        OR p.author = ?     -- User's own posts
        OR (p.privacy = 2 AND f.follower_id IS NOT NULL)  -- Posts visible to followers
        OR (p.privacy = 3 AND pv.user_id IS NOT NULL)    -- Private posts user can see
        ORDER BY p.created_at DESC`,
		userID, userID, userID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch posts",
		})
		log.Printf("Error fetching posts: %v", err)
		return
	}
	defer rows.Close()

	var posts []m.Post
	for rows.Next() {
		var post m.Post
		var groupID sql.NullInt64 // Handle nullable group_id

		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Media,
			&post.Privacy,
			&post.Author,
			&post.CreatedAt,
			&groupID,
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Error reading posts",
			})
			log.Printf("Error scanning post: %v", err)
			return
		}

		// Fetch author information
		var authorName string
		var authorAvatar string
		err = sqlite.DB.QueryRow(`
            SELECT username, avatar
            FROM users
            WHERE id = ?`,
			post.Author).Scan(&authorName, &authorAvatar)
		if err != nil {
			continue // Skip this post if we can't get author info
		}

		post.AuthorName = authorName
		post.AuthorAvatar = authorAvatar

		// Handle nullable group_id
		if groupID.Valid {
			post.GroupID = int(groupID.Int64)
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error iterating through posts",
		})
		return
	}

	json.NewEncoder(w).Encode(posts)
}

// Add this helper function
func canUserViewPost(postID, userID int) (bool, error) {
	var post m.Post
	var isAllowed bool

	// First get the post details
	err := sqlite.DB.QueryRow(`
        SELECT p.privacy, p.author 
        FROM posts p 
        WHERE p.id = ?`, postID).Scan(&post.Privacy, &post.Author)
	if err != nil {
		return false, err
	}

	// If user is the author or post is public, they can view it
	if post.Author == userID || post.Privacy == 1 {
		return true, nil
	}

	// For almost-private posts (followers only)
	if post.Privacy == 2 {
		err = sqlite.DB.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM followers 
                WHERE follower_id = ? AND followed_id = ?
            )`, userID, post.Author).Scan(&isAllowed)
		return isAllowed, err
	}

	// For private posts
	if post.Privacy == 3 {
		err = sqlite.DB.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM post_PrivateViews 
                WHERE post_id = ? AND user_id = ?
            )`, postID, userID).Scan(&isAllowed)
		return isAllowed, err
	}

	return false, nil
}

func ViewPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get post ID from URL
	postIDStr := r.PathValue("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID < 1 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid post ID",
		})
		return
	}

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user information",
		})
		return
	}

	// Fetch the post
	var post m.Post
	err = sqlite.DB.QueryRow(`
		SELECT p.id, p.title, p.content, p.media, p.privacy, p.author, p.created_at, p.group_id
		FROM posts p
		WHERE p.id = ?`,
		postID).Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privacy, &post.Author, &post.CreatedAt, &post.GroupID)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Post not found",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch post",
		})
		return
	}

	// Check if user has permission to view the post
	if post.Privacy != 1 && post.Author != userID {
		// For private posts, check if user is a follower
		if post.Privacy == 3 {
			var isFollower bool
			err = sqlite.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM followers 
					WHERE follower_id = ? AND following_id = ?
				)`,
				userID, post.Author).Scan(&isFollower)
			if err != nil || !isFollower {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "You don't have permission to view this post",
				})
				return
			}
		} else {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "You don't have permission to view this post",
			})
			return
		}
	}

	// Return the post
	json.NewEncoder(w).Encode(post)
}
