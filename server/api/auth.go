package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type header first
	// Set content type header first

	w.Header().Set("Content-Type", "application/json")

	// Create a struct to receive the raw JSON
	type RegisterRequest struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		Username    string `json:"username"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		DateOfBirth string `json:"date_of_birth"`
		Avatar      string `json:"avatar"`
		AboutMe     string `json:"about_me"`
	}

	var rawData RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON format",
		})
		return
	}

	// Parse the date string
	dateOfBirth, err := time.Parse(time.RFC3339, rawData.DateOfBirth)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid date format. Please use ISO 8601 format",
		})
		return
	}

	// Create the user object
	user := m.User{
		Email:       rawData.Email,
		Password:    rawData.Password,
		Username:    rawData.Username,
		FirstName:   rawData.FirstName,
		LastName:    rawData.LastName,
		DateOfBirth: &dateOfBirth,
		Avatar:      rawData.Avatar,
		AboutMe:     rawData.AboutMe,
	}

	// Debug log
	log.Printf("Received registration data: %+v", user)

	// check only required fields
	if strings.TrimSpace(user.Email) == "" ||
		strings.TrimSpace(user.Username) == "" ||
		strings.TrimSpace(user.Password) == "" ||
		strings.TrimSpace(user.FirstName) == "" ||
		strings.TrimSpace(user.LastName) == "" {
		log.Printf("Missing required fields: email=%s, username=%s, firstName=%s, lastName=%s",
			user.Email, user.Username, user.FirstName, user.LastName)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Please provide all required fields: email, username, password, firstName, and lastName",
		})
		return
	}

	var id int
	var err2 error // Declare new error variable
	// check if the username or email already exists
	err2 = sqlite.DB.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", user.Email, user.Username).Scan(&id)
	if err2 == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "User already exists",
		})
		return
	} else if err2 != sql.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Something went wrong",
		})
		log.Printf("select err: %v", err2)
		return
	}

	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error processing password",
		})
		log.Printf("hash error: %v", err)
		return
	}

	// If date of birth is not provided, use current time
	if user.DateOfBirth.IsZero() {
		user.DateOfBirth = &time.Time{}
		*user.DateOfBirth = time.Now()
	}

	res, err := sqlite.DB.Exec(`
		INSERT INTO users (
			username, email, password, first_name, last_name, 
			avatar, about_me, date_of_birth
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Username, user.Email, string(hashedpassword),
		user.FirstName, user.LastName, user.Avatar,
		user.AboutMe, user.DateOfBirth)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create user",
		})
		log.Printf("Insert error: %v", err)
		return
	}

	userID, err := res.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user ID",
		})
		log.Printf("get id error: %v", err)
		return
	}

	// generate the session for the user
	if err := util.GenerateSession(w, &user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create session",
		})
		log.Printf("Session error: %v", err)
		return
	}

	// Create response
	response := map[string]interface{}{
		"id":       userID,
		"username": user.Username,
		"status":   "success",
		"message":  "Registration successful",
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
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

	// Add debug logging
	log.Printf("Cookies received: %v", r.Cookies())

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("Session error: %v", err)
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
			log.Printf("User not found: %s", username)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "User not found",
			})
			return
		}
		log.Printf("Database error: %v", err)
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
