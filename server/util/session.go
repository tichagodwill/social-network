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
	// generating the session id
	sessionID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	sessionIDInString := sessionID.String()

	// create the cookie
	cookie := &http.Cookie{
		Name:     "AccessToken",
		Value:    sessionIDInString,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(24 * time.Hour / time.Second),
		Domain:   "localhost",
	}

	// send the cookie to the browser
	http.SetCookie(w, cookie)

	UserSession[sessionIDInString] = u.Username
	log.Printf("Created session for user: %s", u.Username)

	return nil
}

func DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("AccessToken")
	if err != nil {
		if err == http.ErrNoCookie {
			// no cookie found nothing to do
			return
		}
		return
	}

	// invalidate the cookie and send it to the frontend
	invalidCookie := &http.Cookie{
		Name:     "AccessToken",
		 Value:    "",
		 Path:     "/",
		 HttpOnly: true,
		 Secure:   false,
		 SameSite: http.SameSiteLaxMode,
		 MaxAge:   -1,
	}

	http.SetCookie(w, invalidCookie)

	// remove the cookie from the map
	delete(UserSession, cookie.Value)
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

	log.Printf("Session found for user: %s", username)
	return username, nil
}

