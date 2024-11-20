package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user m.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Error reading data", http.StatusBadRequest)
		return
	}

	// Debug log
	log.Printf("Received registration data: %+v", user)

	// check if all the required fields are provided
	if strings.TrimSpace(user.Email) == "" || strings.TrimSpace(user.Username) == "" || strings.TrimSpace(user.Password) == "" || strings.TrimSpace(user.FirstName) == "" || strings.TrimSpace(user.LastName) == "" || strings.TrimSpace(user.AboutMe) == "" || strings.TrimSpace(user.Avatar) == "" || user.DateOfBirth.IsZero() {
		log.Printf("Missing required fields: email=%s, username=%s, firstName=%s, lastName=%s, aboutMe=%s, avatar=%s, dateOfBirth=%v",
			user.Email, user.Username, user.FirstName, user.LastName, user.AboutMe, user.Avatar, user.DateOfBirth)
		http.Error(w, "Please populate all required fields", http.StatusBadRequest)
		return
	}

	var id int

	// check if the username or email already exists
	err := sqlite.DB.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", user.Email, user.Username).Scan(&id)

	if err == nil {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("select err: %v", err)
		return
	}

	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		log.Printf("hash error: %v", err)
		return
	}

	res, err := sqlite.DB.Exec("INSERT INTO users (username, email, password, first_name, last_name, avatar, about_me, date_of_birth) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", user.Username, user.Email, string(hashedpassword), user.FirstName, user.LastName, user.Avatar, user.AboutMe, user.DateOfBirth)
	if err != nil {
		http.Error(w, "Something went wrong, please try again later", http.StatusInternalServerError)
		log.Printf("Hash error: %v", err)
		return
	}

	// get the last inserted id from the database
	userID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("get id: %v", err)
		return
	}

	// generate the session for the user
	if err := util.GenerateSession(w, &user); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	userResponse := m.UserResponse{
		ID:       userID,
		Username: user.Username,
	}

	if err := json.NewEncoder(w).Encode(&userResponse); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)

		log.Printf("sending back: %v", err)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Always set content type header first
	w.Header().Set("Content-Type", "application/json")

	var loginRequest m.User
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error reading data",
		})
		return
	}

	// both can't be empty one has to be populated
	if strings.TrimSpace(loginRequest.Username) == "" && strings.TrimSpace(loginRequest.Email) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Provide a valid identifier",
		})
		return
	}

	// get the user from the database
	var user m.User
	err := sqlite.DB.QueryRow(`
		SELECT id, username, email, password, first_name, last_name, avatar, about_me, is_private, date_of_birth 
		FROM users 
		WHERE email = ? OR username = ?`,
		loginRequest.Email, loginRequest.Username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.FirstName, &user.LastName, &user.Avatar,
		&user.AboutMe, &user.IsPrivate, &user.DateOfBirth)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "User does not exist",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Something went wrong",
		})
		log.Printf("Login: %v", err)
		return
	}

	// compare the passed password with the existing one
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Username or Password incorrect",
		})
		return
	}

	// generate the session for the user
	if err := util.GenerateSession(w, &user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create session",
		})
		return
	}

	// Create response without sensitive data
	response := map[string]interface{}{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"firstName":   user.FirstName,
		"lastName":    user.LastName,
		"avatar":      user.Avatar,
		"aboutMe":     user.AboutMe,
		"isPrivate":   user.IsPrivate,
		"dateOfBirth": user.DateOfBirth.Format("2006-01-02"),
		"status":      "success",
		"message":     "Login successful",
	}

	// Set status code before writing response
	w.WriteHeader(http.StatusOK)

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		// At this point, we can't write another status code because headers are already sent
		// Just log the error and return
		log.Printf("Failed to encode response: %v", err)
		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type header first
	w.Header().Set("Content-Type", "application/json")

	util.DestroySession(w, r)

	// Set status code
	w.WriteHeader(http.StatusOK)

	// Send JSON response
	response := map[string]string{
		"status":  "success",
		"message": "User logged out successfully",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}

func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Set content type header first
	w.Header().Set("Content-Type", "application/json")

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	var user m.User
	err = sqlite.DB.QueryRow(`
		SELECT id, username, email, first_name, last_name, avatar, about_me, is_private, date_of_birth 
		FROM users 
		WHERE username = ?`, username).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.FirstName, &user.LastName, &user.Avatar,
		&user.AboutMe, &user.IsPrivate, &user.DateOfBirth)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "User not found",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Something went wrong",
		})
		return
	}

	// Create response without sensitive data
	response := map[string]interface{}{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"firstName":   user.FirstName,
		"lastName":    user.LastName,
		"avatar":      user.Avatar,
		"aboutMe":     user.AboutMe,
		"isPrivate":   user.IsPrivate,
		"dateOfBirth": user.DateOfBirth.Format("2006-01-02"),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
