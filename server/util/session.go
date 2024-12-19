package util

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"fmt"
	m "social-network/models"
	"log"
)

var UserSession = make(map[string]string) // sessionID: username

func GenerateSession(w http.ResponseWriter, u *m.User) error {
	sessionID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	sessionIDString := sessionID.String()

	cookie := &http.Cookie{
		Name:     "AccessToken",
		Value:    sessionIDString,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(24 * time.Hour.Seconds()),
		Domain:   "localhost",
	}

	http.SetCookie(w, cookie)
	UserSession[sessionIDString] = u.Username
	log.Printf("Created session for user: %s with token: %s", u.Username, sessionIDString)

	return nil
}

func GetUsernameFromSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie("AccessToken")
	if err != nil {
		return "", fmt.Errorf("no session cookie found: %v", err)
	}

	username, ok := UserSession[cookie.Value]
	if !ok {
		return "", fmt.Errorf("invalid or expired session")
	}

	log.Printf("Session found for user: %s with token: %s", username, cookie.Value)
	return username, nil
}

func DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("AccessToken")
	if err == nil {
		delete(UserSession, cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "AccessToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Domain:   "localhost",
	})
}

