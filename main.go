package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var tmpl *template.Template

//go:embed templates/*
var tmpls embed.FS

type Data struct {
	Title string
	Body  string
}

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	t, err := template.ParseFS(tmpls, "templates/*.html")
	if err != nil {
		log.Printf("error with parsing html templates: %v", err)
	}

	tmpl = t

	r.Get("/", handlerIndex)
	r.Get("/up/", handlerKamalHealthCheck)

	http.ListenAndServe("0.0.0.0:8080", r)
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	data := Data{}

	data.Title = "Stop sucking"
	data.Body = "I am the body"

	tmpl.ExecuteTemplate(w, "index", data)

}

func handlerKamalHealthCheck(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
}
