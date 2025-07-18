package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	// Db open and setup. Create tables if they don't exist
	err := openDb()
	if err != nil {
		log.Fatal(err)
	}
	defer closeDb()

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

	// staticSubFS, _ := fs.Sub(staticFiles, "static")
	// FileServer(r, "/css", http.FS(staticSubFS))
	workDir, _ := os.Getwd()
	staticDir := http.Dir(filepath.Join(workDir, "static"))
	FileServer(r, "/static", staticDir)

	// staticDir := http.Dir(filepath.Join(".", "static"))
	// r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticDir)))

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "login_page.html", providerIndex)
	})
	r.Get("/up/", apiCfg.handlerKamalHealthcheck)
	r.Get("/logout/{provider}", apiCfg.handlerLogout)
	r.Get("/auth/{provider}", apiCfg.handlerGetAuth)
	r.Get("/auth/{provider}/callback", apiCfg.handlerGetAuthCallback)
	r.Get("/profile", apiCfg.handlerProfile)

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

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
