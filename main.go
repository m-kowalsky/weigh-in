package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/m-kowalsky/weigh-in/internal/auth"
	"github.com/m-kowalsky/weigh-in/internal/database"
	"github.com/markbates/goth/gothic"
	_ "github.com/mattn/go-sqlite3"
)

var tmpl *template.Template

//go:embed templates/*
var tmpls embed.FS

type Data struct {
	Title string
	Body  string
}

const tmpl_path = "templates/*.html"

type apiConfig struct {
	db           *database.Queries
	access_token string
}

type User struct {
	Id        string
	Email     string
	Firstname string
}

func main() {

	providerIndex := gothProviderSetup()
	auth.NewAuth()
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	err := openDb()
	if err != nil {
		log.Fatal(err)
	}
	defer closeDb()

	err = setupDbSchema()
	if err != nil {
		log.Fatal(err)
	}

	err = parseHTMLTemplates(tmpl_path)
	if err != nil {
		log.Fatal(err)
	}
	db_queries := database.New(Db)

	apiCfg := apiConfig{
		db:           db_queries,
		access_token: "",
	}

	// Routes

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		err = tmpl.ExecuteTemplate(w, "index.html", providerIndex)
		if err != nil {
			log.Fatal(err)
		}

	})
	r.Get("/up/", apiCfg.handlerKamalHealthcheck)
	r.Get("/logout/{provider}", logoutProviderFunction)
	r.Get("/auth/{provider}", handlerGetAuthProvider)
	r.Get("/auth/{provider}/callback", handlerGetAuthCallback)

	// Run server

	err = http.ListenAndServe(":8080", r)
	log.Println("Serving on port 8080....")
	if err != nil {
		log.Fatal(err)
	}
}

func parseHTMLTemplates(path string) error {

	t, err := template.ParseFS(tmpls, path)
	if err != nil {
		return err
	}

	tmpl = t

	return nil

}

func (apiCfg *apiConfig) handlerKamalHealthcheck(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(http.StatusOK)
}

func handlerGetAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Println("Auth error:", err)
		fmt.Fprintln(w, err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "user.html", user)

	fmt.Println(user.UserID)
	fmt.Println(user.Email)
	fmt.Println(user.FirstName)
	fmt.Printf("type of access token: %T", user.AccessToken)
	http.Redirect(w, r, "/", http.StatusFound)
}

func handlerGetAuthProvider(w http.ResponseWriter, r *http.Request) {

	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		err := tmpl.ExecuteTemplate(w, "user.html", gothUser)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func logoutProviderFunction(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

func gothProviderSetup() *ProviderIndex {

	m := map[string]string{
		"google": "Google",
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}

	return providerIndex
}
