package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post m.Post

	// check if the user is logged in
	if err := util.ValidateSession(r); err != nil {
		if strings.Contains(err.Error(), "invalid session") || strings.Contains(err.Error(), "token does not exists") {
			http.Error(w, "User Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Error reading data", http.StatusBadRequest)
		return
	}

	if _, err := sqlite.DB.Exec("INSERT INTO post (title, content, media, privacy, author) VALUES (?, ?, ?, ?, ?)", post.Title, post.Content, post.Media, post.Privay, post.Author); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Fatalf("create post: %v", err)
		return
	}

	w.Write([]byte("Post created successfully"))
}
