package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/m-kowalsky/weigh-in/internal/database"
	"github.com/markbates/goth/gothic"
)

func (apiCfg *apiConfig) handlerKamalHealthcheck(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(http.StatusOK)
}

func (apiCfg *apiConfig) handlerGetAuthCallback(w http.ResponseWriter, r *http.Request) {

	// Get provider param from url for gothic auth
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	// client_id := os.Getenv("GOOGLE_CLIENT_ID")
	//
	// fmt.Printf("\n client id: %v", client_id)
	goth_user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Println("Auth error:", err)
		fmt.Fprintln(w, err)
		return
	}

	// Check if user exists in db already by getting a count of a user by email
	count, err := apiCfg.db.CheckIfUserExistsByEmail(r.Context(), goth_user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if count == 1 {
		current_user, err := apiCfg.db.GetUserByEmail(r.Context(), goth_user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("current user: %v", current_user)
	} else {
		new_user, err := apiCfg.db.CreateUser(r.Context(), database.CreateUserParams{
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Email:       goth_user.Email,
			AccessToken: goth_user.AccessToken,
			FullName:    goth_user.Name,
		})
		if err != nil {
			http.Error(w, "problem creating new user", http.StatusBadRequest)
		}

		fmt.Printf("current user: %v", new_user)
	}

	// Create new session with user id and email

	sess, err := gothic.Store.Get(r, "user_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	sess.Values["user_email"] = goth_user.Email
	sess.Values["user_id"] = goth_user.UserID
	sess.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)
}

func (apiCfg *apiConfig) handlerGetAuth(w http.ResponseWriter, r *http.Request) {

	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		tmpl.ExecuteTemplate(w, "profile.html", gothUser)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func (apiCfg *apiConfig) handlerLogout(w http.ResponseWriter, r *http.Request) {

	// Logout a user by setting the current user_session max age to -1 which will cause the client to delete the cookie associated with the session
	session, err := gothic.Store.Get(r, "user_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	session.Options.MaxAge = -1
	session.Values = make(map[any]any)
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (apiCfg *apiConfig) handlerProfile(w http.ResponseWriter, r *http.Request) {

	sess, _ := gothic.Store.Get(r, "user_session")
	if sess.IsNew == true {
		fmt.Println(sess.Options.MaxAge)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	email := sess.Values["user_email"].(string)
	fmt.Printf("email from profile: %v", email)
	current_user, err := apiCfg.db.GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	fmt.Printf("\ncurrent user: %s\n", current_user.FullName)

	type ProfileData struct {
		User      database.User
		Providers []string
	}

	data := ProfileData{
		User:      current_user,
		Providers: apiCfg.providerIndex.Providers,
	}

	tmpl.ExecuteTemplate(w, "profile.html", data)

}
