package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

var templates *template.Template

func InitTemplates() error {
	funcMap := template.FuncMap{
		"formatDate": func(t interface{}) string {
			switch v := t.(type) {
			case time.Time:
				return v.Format("Jan 2, 2006 3:04 PM")
			case *time.Time:
				if v != nil {
					return v.Format("Jan 2, 2006 3:04 PM")
				}
				return ""
			default:
				return ""
			}
		},
	}

	var err error
	templates = template.Must(
		template.New("").Funcs(funcMap).ParseGlob("templates/*.html"),
	)
	templates = template.Must(templates.ParseGlob("templates/partials/*.html"))

	log.Printf("Templates loaded successfully")
	return err
}

func Render(w http.ResponseWriter, name string, data interface{}) error {
	log.Printf("Rendering template: %s", name)
	err := templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("Template render error: %v", err)
	}
	return err
}
