package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/shindakun/attodo/internal/config"
	"github.com/shindakun/attodo/internal/version"
)

var templates *template.Template
var appConfig *config.Config

func InitTemplates(cfg *config.Config) error {
	appConfig = cfg

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
		"getVersion": func() string {
			return version.GetVersion()
		},
		"getCommitID": func() string {
			return version.GetCommitID()
		},
		"getBaseURL": func() string {
			if appConfig != nil {
				return appConfig.BaseURL
			}
			return ""
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
