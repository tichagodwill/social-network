package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post m.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Error reading data", http.StatusBadRequest)
		return
	}

	// check if the passed privacy is within the allowed range
	if post.Privay != 1 && post.Privay != 2 && post.Privay != 3 {
		http.Error(w, "invalid privacy type", http.StatusBadRequest)
		return
	}

	if _, err := sqlite.DB.Exec("INSERT INTO posts (title, content, media, privacy, author) VALUES (?, ?, ?, ?, ?)", post.Title, post.Content, post.Media, post.Privay, post.Author); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("create post: %v", err)
		return
	}

	w.Write([]byte("Post created successfully"))
}

// the handler that contains the logic for viewing the post
func ViewPost(w http.ResponseWriter, r *http.Request) {

	// get the id from the path
	idString := r.PathValue("id")

	// convert the id into number
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	var post m.Post
	if err := sqlite.DB.QueryRow("SELECT * FROM posts WHERE id = ?", id).Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privay, &post.Author, &post.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post does not exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	if err := json.NewEncoder(w).Encode(&post); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	var posts []m.Post

	row, err := sqlite.DB.Query("SELECT * FROM posts WHERE privacy = 1 AND group = NULL ")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	// go through all the posts
	for row.Next() {
		var post m.Post

		// get individual post and copy the values into the variable
		if err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privay, &post.Author, &post.CreatedAt); err != nil {
			http.Error(w, "Error getting post", http.StatusInternalServerError)
			log.Printf("Error scanning: %v", err)
			return
		}

		// append the post to the slice
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	// send the array of posts to the frontend
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
}
