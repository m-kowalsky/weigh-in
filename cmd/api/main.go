package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var tmpl *template.Template

type Data struct {
	Title string
	Body  string
}

func main() {

	r := chi.NewRouter()

	t, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Printf("error with parsing html templates: %v", err)
	}

	tmpl = t

	r.Get("/", handlerIndex)

	http.ListenAndServe(":8080", r)
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	data := Data{}

	data.Title = "Stop sucking"
	data.Body = "I am the body"

	tmpl.ExecuteTemplate(w, "index", data)

}
