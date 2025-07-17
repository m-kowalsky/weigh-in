package auth

import (
	"fmt"
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
	fmt.Printf("\n client id, client secret: %v, %v\n", googleClientId, googleClientSecret)

	store := sessions.NewCookieStore([]byte(session_key))
	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	var callback_url string
	if IsProd {
		callback_url = "https://www.mdksoftware.io/auth/google/callback"
	} else {
		callback_url = "http://localhost:8080/auth/google/callback"
	}

	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, callback_url),
	)
	fmt.Println(callback_url)
}
