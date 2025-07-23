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

//go:embed templates/*
var tmpls embed.FS

//go:embed sql/schema/*
var migrations embed.FS

var tmpl *template.Template

type Data struct {
	Title string
	Body  string
}

const session_name = "user-session"

const tmpl_path = "templates/*.html"

type apiConfig struct {
	db            *database.Queries
	providerIndex *ProviderIndex
}

func main() {

	// Goth and sessions setup
	providerIndex := gothProviderSetup()
	auth.NewAuth()

	// Setup chi router and logger middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	if auth.IsProd {
		// Db open and setup. Create tables if they don't exist
		err := openDb("/app/data/app.db")
		if err != nil {
			log.Fatal(err)
		}
		defer closeDb()
	} else {
		err := openDb("./dev_test.db")
		if err != nil {
			log.Fatal(err)
		}
		defer closeDb()
	}

	// Goose migrations setup
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal("Error setting goose dialect")
	}

	if err := goose.Up(Db, "sql/schema"); err != nil {
		log.Fatal("Error running database migrations")
	}

	// Parse templates in /templates/*.html
	err := parseHTMLTemplates(tmpl_path)
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
		tmpl.ExecuteTemplate(w, "login_page.html", providerIndex)
	})
	r.Get("/up/", apiCfg.handlerKamalHealthcheck)
	r.Get("/logout/{provider}", apiCfg.handlerLogout)
	r.Get("/auth/{provider}", apiCfg.handlerGetAuth)
	r.Get("/auth/{provider}/callback", apiCfg.handlerGetAuthCallback)
	r.Get("/profile", apiCfg.handlerProfile)
	r.Post("/user/{user_id}", apiCfg.handlerGetUser)

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

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
