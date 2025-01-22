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

	// Insert the post into the database
	result, err := sqlite.DB.Exec(
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

	// Return success response with post ID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Post created successfully",
		"id":      postID,
	})
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

	// Check the database connection before running the query.
	if sqlite.DB == nil {
		log.Printf("Error: Database is not initialized")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Database not initialized",
		})
		return
	}
	err = sqlite.DB.Ping() // Check the connection
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Database connection failed",
		})
		return
	}
	log.Printf("UserID: %v", userID)

	// Fetch posts from the database
	rows, err := sqlite.DB.Query(`
		SELECT p.id, p.title, p.content, p.media, p.privacy, p.author, p.created_at, p.group_id
		FROM posts p
		LEFT JOIN followers f ON p.author = f.followed_id
		WHERE p.privacy = 1  -- Public posts
		OR p.author = ?     -- User's own posts
		OR (p.privacy = 3 AND f.follower_id = ?)  -- Posts visible to followers
		ORDER BY p.created_at DESC`,
		userID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch posts",
		})
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
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Error reading posts",
			})
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
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to fetch author's username",
			})
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
