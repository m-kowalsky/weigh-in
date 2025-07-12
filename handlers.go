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
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

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

		apiCfg.current_user = current_user
		fmt.Printf("current user: %v", apiCfg.current_user)
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

		apiCfg.current_user = new_user
		fmt.Printf("current user: %v", apiCfg.current_user)
	}

	sess, err := gothic.Store.Get(r, "gothic-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	sess.Values["is_auth"] = true
	sess.Values["user_email"] = apiCfg.current_user.Email
	sess.Values["user_id"] = apiCfg.current_user.ID
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
	sess, _ := gothic.Store.Get(r, "gothic-session")
	sess.Values["is_auth"] = false
	sess.Options.MaxAge = -1
	sess.Save(r, w)

	gothic.Logout(w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (apiCfg *apiConfig) handlerProfile(w http.ResponseWriter, r *http.Request) {

	if !IsAuth(r) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return

	}

	user_email := apiCfg.current_user.Email

	fmt.Printf("User email: %s", user_email)

	data := struct {
		Providers []string
		Email     string
	}{
		Providers: apiCfg.providerIndex.Providers,
		Email:     user_email,
	}

	tmpl.ExecuteTemplate(w, "profile.html", data)

}

func IsAuth(r *http.Request) bool {
	sess, _ := gothic.Store.Get(r, "gothic-session")

	// Check "authenticated" flag
	if auth, ok := sess.Values["is_auth"].(bool); !ok || !auth {
		// Not logged in → redirect to home or login
		return false
	}
	return true

}
