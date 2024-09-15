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
)

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var group m.Group

	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(group.Description) == "" || strings.TrimSpace(group.Title) == "" {
		http.Error(w, "Please Provide all fields", http.StatusBadRequest)
		return
	}

	// check if the user exists
	if exist := m.DoesUserExist(group.CreatorID, sqlite.DB); !exist {
		http.Error(w, "User does not exists", http.StatusBadRequest)
		return
	}

	if _, err := sqlite.DB.Exec("INSERT INTO groups (creator_id, title, description) VALUES (?, ?, ?)", group.CreatorID, group.Title, group.Description); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	w.Write([]byte("Group created successfully"))
}

func CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	var post m.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	groupIDString := r.PathValue("id")

	// convert the string into a number
	groupID, err := strconv.Atoi(groupIDString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}
	// post will always be public for the group members
	post.Privay = 1

	// check if the passed privacy is within the allowed range
	if post.Privay != 1 && post.Privay != 2 && post.Privay != 3 {
		http.Error(w, "invalid privacy type", http.StatusBadRequest)
		return
	}

	if _, err := sqlite.DB.Exec("INSERT INTO posts (title, content, media, privacy, author, group_id) VALUES (?, ?, ?, ?, ?, ?)", post.Title, post.Content, post.Media, post.Privay, post.Author, groupID); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("create post: %v", err)
		return
	}

	w.Write([]byte("Post created successfully"))
}

func GetGroupPost(w http.ResponseWriter, r *http.Request) {
	var groupPosts []m.Post
	groupIDString := r.PathValue("id")

	// convert the string into a number
	groupID, err := strconv.Atoi(groupIDString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	// the value of the group id can't be less then 1
	if groupID < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	rows, err := sqlite.DB.Query("SELECT * FROM posts WHERE group_id = ?", groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Group does not exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	for rows.Next() {
		var post m.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privay, &post.Author, &post.CreatedAt); err != nil {
			http.Error(w, "Error getting post", http.StatusInternalServerError)
			log.Printf("Error scanning: %v", err)
			return
		}

		groupPosts = append(groupPosts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&groupPosts); err != nil {
		http.Error(w, "Error sending json", http.StatusInternalServerError)
	}
}
