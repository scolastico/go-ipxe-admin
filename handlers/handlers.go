package handlers

import (
	"bytes"
	"embed"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/scolastico/go-ipxe-admin/models"
)

type Handlers struct{}

func New() *Handlers {
	return &Handlers{}
}

func StaticHandler(content embed.FS) http.Handler {
	return http.FileServer(http.FS(content))
}

func (h *Handlers) Dashboard(c echo.Context) error {
	return c.Render(http.StatusOK, "dashboard.html", nil)
}

func (h *Handlers) ServeIPXE(c echo.Context) error {
	ip := c.RealIP()

	// Default fallbacks if no IP header (for local testing)
	if ip == "" || ip == "::1" || ip == "127.0.0.1" {
		ip = c.QueryParam("ip")
	}

	if ip != "" {
		models.TrackClient(ip)
	}

	deps, _ := models.LoadDeployments()
	var d *models.Deployment
	for _, dep := range deps {
		if dep.IP == ip {
			depCopy := dep
			d = &depCopy
			break
		}
	}

	if d == nil {
		for _, dep := range deps {
			if strings.Contains(dep.IP, "/") {
				_, ipnet, err := net.ParseCIDR(dep.IP)
				if err == nil {
					cip := net.ParseIP(ip)
					if cip != nil && ipnet.Contains(cip) {
						depCopy := dep
						d = &depCopy
						break
					}
				}
			}
		}
	}

	if d == nil {
		for _, dep := range deps {
			if dep.IP == "" {
				depCopy := dep
				d = &depCopy
				break
			}
		}
	}

	if d == nil {
		return c.String(http.StatusNotFound, "#!ipxe\necho No deployment found for IP "+ip+"\nsleep 3\nexit 1")
	}

	if d.Timeout > 0 {
		createdAt, err := time.Parse(time.RFC3339, d.CreatedAt)
		if err == nil && time.Since(createdAt) > time.Duration(d.Timeout)*time.Minute {
			models.DeleteDeployment(d.IP)
			return c.String(http.StatusNotFound, "#!ipxe\necho Deployment expired\nshell")
		}
	}

	tmpls, _ := models.LoadTemplates()
	var t *models.Template
	for _, tmpl := range tmpls {
		if tmpl.ID == d.TemplateID {
			t = &tmpl
			break
		}
	}

	if t == nil {
		return c.String(http.StatusNotFound, "#!ipxe\necho Template not found\nshell")
	}

	// Simple template replacement
	parsedTmpl, err := template.New("ipxe").Parse(t.Content)
	if err != nil {
		return c.String(http.StatusInternalServerError, "#!ipxe\necho Template parse error\nshell")
	}

	var buf bytes.Buffer
	err = parsedTmpl.Execute(&buf, d.Variables)
	if err != nil {
		return c.String(http.StatusInternalServerError, "#!ipxe\necho Template execute error\nshell")
	}

	content := buf.String()
	if !strings.HasPrefix(strings.TrimSpace(content), "#!ipxe") {
		content = "#!ipxe\n" + content
	}

	if d.Once {
		models.DeleteDeployment(d.IP)
	}

	return c.String(http.StatusOK, content)
}

func (h *Handlers) ServeAsset(c echo.Context) error {
	id := c.Param("id")
	filename := c.Param("filename")

	return c.File(filepath.Join("data", "assets", id+"_"+filename))
}

func (h *Handlers) GetDeployments(c echo.Context) error {
	deps, _ := models.LoadDeployments()
	if deps == nil {
		deps = []models.Deployment{}
	}
	return c.JSON(http.StatusOK, deps)
}

func (h *Handlers) SaveDeployment(c echo.Context) error {
	var d models.Deployment
	if err := c.Bind(&d); err != nil {
		return err
	}
	d.CreatedAt = time.Now().Format(time.RFC3339)
	models.SaveDeployment(d)
	return c.JSON(http.StatusOK, d)
}

func (h *Handlers) DeleteDeployment(c echo.Context) error {
	ip := c.Param("ip")
	if ip == "__all__" {
		ip = ""
	}
	models.DeleteDeployment(ip)
	return c.NoContent(http.StatusOK)
}

func (h *Handlers) GetRecentClients(c echo.Context) error {
	clients, _ := models.GetRecentClients(30)
	if clients == nil {
		clients = []models.Client{}
	}
	return c.JSON(http.StatusOK, clients)
}

func (h *Handlers) GetTemplates(c echo.Context) error {
	tmpls, _ := models.LoadTemplates()
	if tmpls == nil {
		tmpls = []models.Template{}
	}
	return c.JSON(http.StatusOK, tmpls)
}

func (h *Handlers) SaveTemplate(c echo.Context) error {
	var t models.Template
	if err := c.Bind(&t); err != nil {
		return err
	}
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	models.SaveTemplate(t)
	return c.JSON(http.StatusOK, t)
}

func (h *Handlers) DeleteTemplate(c echo.Context) error {
	id := c.Param("id")
	models.DeleteTemplate(id)
	return c.NoContent(http.StatusOK)
}

func (h *Handlers) GetAssets(c echo.Context) error {
	assets, _ := models.LoadAssets()
	if assets == nil {
		assets = []models.Asset{}
	}
	return c.JSON(http.StatusOK, assets)
}

func (h *Handlers) SaveAsset(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	id := uuid.New().String()
	filename := file.Filename

	dst, err := os.Create(filepath.Join("data", "assets", id+"_"+filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	a := models.Asset{
		ID:       id,
		Filename: filename,
		MimeType: file.Header.Get("Content-Type"),
	}

	models.SaveAssetMeta(a)

	return c.JSON(http.StatusOK, a)
}

func (h *Handlers) DeleteAsset(c echo.Context) error {
	id := c.Param("id")
	models.DeleteAsset(id)
	return c.NoContent(http.StatusOK)
}

type TextAssetRequest struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

func (h *Handlers) SaveAssetText(c echo.Context) error {
	var req TextAssetRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	id := req.ID
	if id == "" {
		id = uuid.New().String()
	} else {
		models.DeleteAsset(id)
	}

	dst, err := os.Create(filepath.Join("data", "assets", id+"_"+req.Filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = dst.WriteString(req.Content); err != nil {
		return err
	}

	a := models.Asset{
		ID:       id,
		Filename: req.Filename,
		MimeType: "text/plain",
	}

	models.SaveAssetMeta(a)

	return c.JSON(http.StatusOK, a)
}
