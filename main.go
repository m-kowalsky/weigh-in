package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func main() {

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

	r.Get("/", handlerIndex)
	r.Get("/up/", handlerKamalHealthCheck)

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

func handlerIndex(w http.ResponseWriter, _ *http.Request) {
	data := Data{}

	data.Title = "Stop sucking"
	data.Body = "I am trying"

	tmpl.ExecuteTemplate(w, "index", data)

}

func handlerKamalHealthCheck(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(http.StatusOK)
}
