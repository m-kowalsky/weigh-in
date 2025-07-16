package auth

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	MaxAge = 86400 * 30
	IsProd = true
)

func NewAuth() {

	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	session_key := os.Getenv("SESSION_SECRET_KEY")
	if session_key == "" {
		log.Fatal("Session secret key is not set or empty.  Check .env file.")
	}

	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	store := sessions.NewCookieStore([]byte(session_key))
	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, "http://localhost:8080/auth/google/callback"),
	)
}
