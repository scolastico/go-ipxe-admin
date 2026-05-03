package main

import (
	"embed"
	"html/template"
	"io"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/scolastico/go-ipxe-admin/handlers"
)

//go:embed templates/* static/*
var content embed.FS

type TemplateRegistry struct {
	templates *template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	// Create required directories
	os.MkdirAll("data/templates", 0755)
	os.MkdirAll("data/assets", 0755)
	os.MkdirAll("data/deployments", 0755)

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Basic Auth for Admin UI
	adminAuth := middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// Simple default auth: admin/admin. In production, read from env.
		adminUser := os.Getenv("ADMIN_USER")
		if adminUser == "" {
			adminUser = "admin"
		}
		adminPass := os.Getenv("ADMIN_PASS")
		if adminPass == "" {
			adminPass = "admin"
		}
		if username == adminUser && password == adminPass {
			return true, nil
		}
		return false, nil
	})

	// Setup templating
	e.Renderer = &TemplateRegistry{
		templates: template.Must(template.ParseFS(content, "templates/*.html")),
	}

	// Setup handlers
	h := handlers.New()

	// Static assets
	e.GET("/static/*", echo.WrapHandler(handlers.StaticHandler(content)))

	// iPXE endpoint
	e.GET("/ipxe", h.ServeIPXE)

	// Assets endpoint
	e.GET("/a/:id/:filename", h.ServeAsset)

	// Admin Dashboard
	admin := e.Group("")
	admin.Use(adminAuth)
	admin.GET("/", h.Dashboard)

	// API routes (also protected by basic auth)
	api := e.Group("/api")
	api.Use(adminAuth)

	api.GET("/deployments", h.GetDeployments)
	api.POST("/deployments", h.SaveDeployment)
	api.DELETE("/deployments/:ip", h.DeleteDeployment)

	api.GET("/clients/recent", h.GetRecentClients)

	api.GET("/templates", h.GetTemplates)
	api.POST("/templates", h.SaveTemplate)
	api.DELETE("/templates/:id", h.DeleteTemplate)

	api.GET("/assets", h.GetAssets)
	api.POST("/assets", h.SaveAsset)
	api.POST("/assets/text", h.SaveAssetText)
	api.DELETE("/assets/:id", h.DeleteAsset)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
