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
		http.Error(w, "Error reading data", 400)
		return
	}

	// check if all the required fields are provided
	if strings.TrimSpace(user.Email) == "" || strings.TrimSpace(user.Username) == "" || strings.TrimSpace(user.Password) == "" || strings.TrimSpace(user.FirstName) == "" || strings.TrimSpace(user.LastName) == "" || user.DateOfBirth.IsZero() {
		http.Error(w, "Please populate all required fields", 400)
		return
	}

	var id int

	// check if the username or email already exists
	err := sqlite.DB.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", user.Email, user.Username).Scan(&id)

	if err == nil {
		http.Error(w, "User already exists", 400)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, "Something went wrong", 500)
		return
	}

	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Something went wrong", 400)
		return
	}

	if _, err := sqlite.DB.Exec("INSERT INTO users (username, email, password, firstname, lastname, date_of_birth) VALUES (?, ?, ?, ?, ?, ?)", user.Username, user.Email, string(hashedpassword), user.FirstName, user.LastName, user.DateOfBirth); err != nil {
		http.Error(w, "Something went wrong, please try again later", 500)
		return
	}

	// generate the session for the user
	if err := util.GenerateSession(w, &user); err != nil {
		http.Error(w, "Something went wrong", 500)
		log.Printf("Error: %v", err)
		return
	}

	if _, err := w.Write([]byte("User regsitered successfully")); err != nil {
		http.Error(w, "Something went wrong", 500)
		log.Printf("Register: %v", err)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user m.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Error reading data", 400)
		return
	}

	// both can't be empty one has to be populated
	if strings.TrimSpace(user.Username) == "" && strings.TrimSpace(user.Email) == "" {
		http.Error(w, "Provide a valid identifier", 400)
		return
	}

	// get the user from the database
	var username, password string
	if err := sqlite.DB.QueryRow("SELECT username, password FROM users WHERE username = ? OR email = ?", user.Username, user.Email).Scan(&username, &password); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User does not exist", 400)
			return
		}
		http.Error(w, "Something went wrong", 500)
		log.Printf("Login: %v", err)
		return
	}

	// compare the passed password with the existing one
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(user.Password)); err != nil {
		http.Error(w, "Username or Password incorrect", 400)
		return
	}

	// generate the session for the user
	util.GenerateSession(w, &user)

	w.Write([]byte("Login successfull"))

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	util.DestroySession(w, r)
	if _, err := w.Write([]byte("User logged out successfully")); err != nil {
		http.Error(w, "Something went wrong", 500)
		log.Printf("Logout: %v", err)
	}

	w.Write([]byte("User logged out successfully"))
}
