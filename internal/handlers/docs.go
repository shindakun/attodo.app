package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/shindakun/attodo/internal/session"
)

// DocsPageData contains data for rendering documentation pages
type DocsPageData struct {
	Avatar     string
	DocTitle   string
	DocContent template.HTML
}

// Docs handles the /docs route and redirects to the index page
func Docs(w http.ResponseWriter, r *http.Request) {
	renderDocsPage(w, r, "index")
}

// DocsPage handles /docs/{name} routes
func DocsPage(w http.ResponseWriter, r *http.Request) {
	// Extract the path after /docs/
	path := strings.TrimPrefix(r.URL.Path, "/docs/")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		path = "index"
	}

	// Security: prevent path traversal
	if strings.Contains(path, "..") || strings.Contains(path, "/") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	renderDocsPage(w, r, path)
}

// renderDocsPage loads and renders a documentation page
func renderDocsPage(w http.ResponseWriter, r *http.Request, docName string) {
	// Load the markdown file
	mdPath := filepath.Join("docs", docName+".md")
	mdContent, err := os.ReadFile(mdPath)
	if err != nil {
		log.Printf("Failed to load doc %s: %v", docName, err)
		http.Error(w, "Documentation not found", http.StatusNotFound)
		return
	}

	// Convert markdown to HTML
	htmlContent := markdownToHTML(mdContent)

	// Get user avatar if logged in
	avatar := ""
	_, ok := session.GetSession(r)
	if ok {
		// For now, we don't have avatar support in attodo, but we check auth
		// This could be extended later if needed
	}

	// Extract title from markdown
	title := extractTitle(string(mdContent))
	if title == "" {
		title = "Documentation"
	}

	// Prepare template data
	data := DocsPageData{
		Avatar:     avatar,
		DocTitle:   title,
		DocContent: template.HTML(htmlContent),
	}

	// Render template
	w.Header().Set("Content-Type", "text/html")
	if err := Render(w, "docs.html", data); err != nil {
		log.Printf("Failed to render docs template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// markdownToHTML converts markdown bytes to HTML
func markdownToHTML(md []byte) []byte {
	// Create parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// Create HTML renderer
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

// extractTitle extracts the first H1 heading from markdown
func extractTitle(md string) string {
	lines := strings.Split(md, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// DocsImage serves images from the docs/images/ directory
func DocsImage(w http.ResponseWriter, r *http.Request) {
	// Extract filename from path
	filename := strings.TrimPrefix(r.URL.Path, "/docs/images/")

	// Security: prevent path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Construct file path
	imgPath := filepath.Join("docs", "images", filename)

	// Check if file exists
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, imgPath)
}
