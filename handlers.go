package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/m-kowalsky/weigh-in/internal/database"
	"github.com/markbates/goth/gothic"
	// "golang.org/x/tools/go/analysis/passes/stringintconv"
)

const sess_email = "user_email"
const sess_userId = "user_id"

func (cfg *apiConfig) handlerKamalHealthcheck(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handlerGetAuthCallback(w http.ResponseWriter, r *http.Request) {

	// Cookie debug
	// for _, cookie := range r.Cookies() {
	// 	fmt.Printf("\nGetAuthCallback-start cookie: name: %v, value: %v\n", cookie.Name, cookie.Value)
	// }

	// Get provider param from url for gothic auth
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	goth_user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Println("Auth error:", err)
		fmt.Fprintln(w, err)
		return
	}

	// Cookie debug
	// for _, cookie := range r.Cookies() {
	// 	fmt.Printf("\nGetAuthCallback-end cookie: name: %v, value: %v\n", cookie.Name, cookie.Value)
	// }

	// Check if user exists in db already by getting a count of a user by email
	count, err := cfg.db.CheckIfUserExistsByEmail(r.Context(), goth_user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if count == 1 {
		current_user, err := cfg.db.GetUserByEmail(r.Context(), goth_user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("current user: %v", current_user)
	} else {
		new_user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
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

	fmt.Printf("\nrefresh token : %v\n expries at: %v\n", goth_user.RefreshToken, goth_user.ExpiresAt)
	// Create new session with user id and email

	sess, err := gothic.Store.Get(r, session_name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	sess.Values[sess_email] = goth_user.Email
	sess.Values[sess_userId] = goth_user.UserID
	sess.AddFlash("Weight Received!", "weight")
	sess.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (cfg *apiConfig) handlerGetAuth(w http.ResponseWriter, r *http.Request) {

	// Cookie debug
	// for _, cookie := range r.Cookies() {
	// 	fmt.Printf("\nGetAuth cookie: name: %v, value: %v\n", cookie.Name, cookie.Value)
	// }

	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		tmpl.ExecuteTemplate(w, "index", gothUser)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func (cfg *apiConfig) handlerLogout(w http.ResponseWriter, r *http.Request) {

	// Logout a user by setting the current user_session max age to -1 which will cause the client to delete the cookie associated with the session
	session, err := gothic.Store.Get(r, session_name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	session.Options.MaxAge = -1
	session.Values = make(map[any]any)
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func (cfg *apiConfig) handlerIndex(w http.ResponseWriter, r *http.Request) {

	sess, _ := gothic.Store.Get(r, session_name)
	if sess.IsNew == true {
		fmt.Println(sess.Options.MaxAge)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	email := sess.Values[sess_email].(string)
	user_id := sess.Values[sess_userId].(string)

	fmt.Printf("email and user_id from session: %v, %v\n", email, user_id)

	current_user, err := cfg.db.GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	fmt.Printf("\ncurrent user: %s\n", current_user.FullName)

	type ProfileData struct {
		User      database.User
		Providers []string
		Title     string
	}

	data := ProfileData{
		User:      current_user,
		Providers: cfg.providerIndex.Providers,
		Title:     "Weigh In",
	}
	fmt.Printf("user from data - index tmpl: %v\n", data.User)

	err = tmpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		fmt.Printf("Template error: %v", err)
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	user_id := chi.URLParam(r, "user_id")

	id_int, err := strconv.ParseInt(user_id, 16, 64)
	if err != nil {
		http.Error(w, "Faile to convert user_id urlParam to int", http.StatusBadRequest)
	}

	current_user, err := cfg.db.GetUserById(r.Context(), id_int)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("user email from handerlGetUser before tmpl execute: %v\n", current_user.Email)

	err = tmpl.ExecuteTemplate(w, "user", current_user)
	if err != nil {
		fmt.Printf("Template error: %v", err)
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
		return
	}
}

// func (cfg *apiConfig) handlerWeighInForm(w http.ResponseWriter, r *http.Request) {
// 	tmpl.ExecuteTemplate(w, "weigh_in_form", nil)
//
// }

func (cfg *apiConfig) handlerLandingPage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "landing_page", nil)

}

func (cfg *apiConfig) handlerCreateWeighIn(w http.ResponseWriter, r *http.Request) {
	sess, _ := gothic.Store.Get(r, session_name)
	if sess.IsNew == true {
		fmt.Println(sess.Options.MaxAge)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	flash_message := sess.Flashes("weight")
	// convert weight string to int64
	// weight, err := strconv.ParseInt(r.FormValue("weight"), 10, 64)
	// if err != nil {
	// 	http.Error(w, "Failed to parse weight to int64", http.StatusBadRequest)
	// }
	//
	// // convert cheated and alcohol string to bool
	// cheated, err := strconv.ParseBool(r.FormValue("cheated"))
	// if err != nil {
	// 	http.Error(w, "Failed to parse cheated to bool", http.StatusBadRequest)
	// }
	// alcohol, err := strconv.ParseBool(r.FormValue("alcohol"))
	// if err != nil {
	// 	http.Error(w, "Failed to parse alcohol to bool", http.StatusBadRequest)
	// }
	//
	// // convert log date string to time.Time
	// time_layout := "2006-01-02"
	// log_date, err := time.Parse(time_layout, r.FormValue("log_date"))
	//
	// new_weigh_in := WeighIn{
	// 	Weight:      weight,
	// 	WeightUnit:  r.FormValue("weight_unit"),
	// 	LogDate:     log_date,
	// 	Cheated:     cheated,
	// 	Alcohol:     alcohol,
	// 	Note:        r.FormValue("note"),
	// 	WeighInDiet: r.FormValue("weigh_in_diet"),
	// }
	fmt.Fprintf(w, "%v", flash_message...)
	// fmt.Printf("log date: %v\n", r.FormValue("log_date"))
	// fmt.Printf("logdate from new weigh in: %v\n", new_weigh_in.LogDate)

	// err = tmpl.ExecuteTemplate(w, "weight", new_weigh_in)
	// if err != nil {
	// 	fmt.Printf("Template error: %v", err)
	// 	http.Error(w, "Template rendering failed", http.StatusInternalServerError)
	// 	return
	// }
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type PageData struct {
		ProviderIndex *ProviderIndex
		Title         string
	}

	data := PageData{
		ProviderIndex: cfg.providerIndex,
		Title:         "Weigh In - Login",
	}
	err := tmpl.ExecuteTemplate(w, "login", data)
	if err != nil {
		fmt.Printf("Template error: %v", err)
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
		return
	}
}
