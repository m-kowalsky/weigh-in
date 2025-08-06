package auth

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	// "github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

const (
	MaxAge = 86400 * 30
	IsProd = false
)

func NewAuth() {

	session_key := os.Getenv("SESSION_SECRET_KEY")
	if session_key == "" {
		log.Fatal("Session secret key is not set or empty.  Check .env file.")
	}

	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	// githubClientId := os.Getenv("GITHUB_CLIENT_ID")
	// githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	store := sessions.NewCookieStore([]byte(session_key))
	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	var google_callback_url string
	if IsProd {
		google_callback_url = "https://mdksoftware.io/auth/google/callback"
	} else {
		google_callback_url = "http://localhost:8080/auth/google/callback"
	}

	// var github_callback_url string
	// if IsProd {
	// 	google_callback_url = "https://mdksoftware.io/auth/github/callback"
	// } else {
	// 	google_callback_url = "http://localhost:8080/auth/github/callback"
	// }

	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, google_callback_url),
		// github.New(githubClientId, githubClientSecret, github_callback_url),
	)
}
