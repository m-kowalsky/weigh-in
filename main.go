package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/m-kowalsky/weigh-in/internal/auth"
	"github.com/m-kowalsky/weigh-in/internal/database"
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

var tmpl *template.Template

//go:embed templates/*
var tmpls embed.FS

//go:embed sql/schema/*
var migrations embed.FS

type Data struct {
	Title string
	Body  string
}

const tmpl_path = "templates/*.html"

type apiConfig struct {
	db            *database.Queries
	access_token  string
	providerIndex *ProviderIndex
}

func main() {

	// Goth and sessions setup
	providerIndex := gothProviderSetup()
	auth.NewAuth()

	// Setup chi router and logger middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Db open and setup. Create tables if they don't exist
	err := openDb()
	if err != nil {
		log.Fatal(err)
	}
	defer closeDb()

	// err = setupDbSchema()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Goose migrations setup
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal("Error setting goose dialect")
	}

	if err := goose.Up(Db, "sql/schema"); err != nil {
		log.Fatal("Error running database migrations")
	}

	// Parse templates in /templates/*.html
	err = parseHTMLTemplates(tmpl_path)
	if err != nil {
		log.Fatal(err)
	}

	// Connect db created above to queries for sqlc
	db_queries := database.New(Db)

	apiCfg := apiConfig{
		db:            db_queries,
		providerIndex: providerIndex,
	}

	// Routes

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "index.html", providerIndex)

	})
	r.Get("/up/", apiCfg.handlerKamalHealthcheck)
	r.Get("/logout/{provider}", apiCfg.handlerLogout)
	r.Get("/auth/{provider}", apiCfg.handlerGetAuth)
	r.Get("/auth/{provider}/callback", apiCfg.handlerGetAuthCallback)
	r.Get("/profile", apiCfg.handlerProfile)

	// Run server

	err = http.ListenAndServe(":80", r)
	log.Println("Serving on port 80....")
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
