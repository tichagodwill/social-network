package util

import (
	"errors"
	"net/http"
	"time"

	"github.com/gofrs/uuid"

	m "social-network/models"
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
		Secure:   true,
		MaxAge:   int(24 * time.Hour / time.Second),
	}

	// send the cookie to the browser
	http.SetCookie(w, cookie)

	UserSession[sessionIDInString] = u.Username

	return nil
}

func ValidateSession(r *http.Request) error {
	// get the cookie from the browser
	cookie, err := r.Cookie("AccessToken")
	if err != nil {
		// check if the cookie exists from the browser
		if err == http.ErrNoCookie {
			return errors.New("token does not exists")
		}
		return err
	}

	// get the value of the cookie
	cookieValue := cookie.Value

	// check if the cookie exists in the already active sessions
	if _, ok := UserSession[cookieValue]; !ok {
		return errors.New("invalid session")
	}

	// session is valid
	return nil
}

func DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("AccessToken")
	if err != nil {
		if err == http.ErrNoCookie {
			// no cookie found nothing to do
			return
		}
		http.Error(w, "Something went wrong", 500)
		return
	}

	// invalidate the cookie and send it to the frontend
	invalidCookie := &http.Cookie{
		Name:     "AccessToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	}

	http.SetCookie(w, invalidCookie)

	cookieValue := cookie.Value
	// remove the cookie from the map
	delete(UserSession, cookieValue)
}
