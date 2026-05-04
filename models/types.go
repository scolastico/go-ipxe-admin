package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type TemplateVariable struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // string, select
	Options string `json:"options"` // comma separated options for select
	Default string `json:"default"`
}

type Template struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Variables []TemplateVariable `json:"variables"`
	Content   string             `json:"content"`
}

type Asset struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
}

type Deployment struct {
	IP         string            `json:"ip"`
	TemplateID string            `json:"template_id"`
	Variables  map[string]string `json:"variables"`
	Once       bool              `json:"once"`
	Timeout    int               `json:"timeout"` // minutes
	CreatedAt  string            `json:"created_at"` // ISO8601
}

type Client struct {
	IP       string `json:"ip"`
	LastSeen string `json:"last_seen"` // time.RFC3339
}

var (
	dataDir string = "data"
	mu      sync.Mutex
)

func SaveTemplate(t Template) error {
	mu.Lock()
	defer mu.Unlock()
	b, _ := json.MarshalIndent(t, "", "  ")
	return os.WriteFile(filepath.Join(dataDir, "templates", t.ID+".json"), b, 0644)
}

func LoadTemplates() ([]Template, error) {
	mu.Lock()
	defer mu.Unlock()
	files, err := os.ReadDir(filepath.Join(dataDir, "templates"))
	if err != nil {
		return nil, err
	}
	var res []Template
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			b, _ := os.ReadFile(filepath.Join(dataDir, "templates", f.Name()))
			var t Template
			json.Unmarshal(b, &t)
			res = append(res, t)
		}
	}
	return res, nil
}

func DeleteTemplate(id string) error {
	mu.Lock()
	defer mu.Unlock()
	return os.Remove(filepath.Join(dataDir, "templates", id+".json"))
}

func safeIP(ip string) string {
	if ip == "" {
		return "_all"
	}
	return strings.ReplaceAll(ip, "/", "_cidr_")
}

func SaveDeployment(d Deployment) error {
	mu.Lock()
	defer mu.Unlock()
	b, _ := json.MarshalIndent(d, "", "  ")
	return os.WriteFile(filepath.Join(dataDir, "deployments", safeIP(d.IP)+".json"), b, 0644)
}

func LoadDeployments() ([]Deployment, error) {
	mu.Lock()
	defer mu.Unlock()
	files, err := os.ReadDir(filepath.Join(dataDir, "deployments"))
	if err != nil {
		return nil, err
	}
	var res []Deployment
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			b, _ := os.ReadFile(filepath.Join(dataDir, "deployments", f.Name()))
			var d Deployment
			json.Unmarshal(b, &d)
			res = append(res, d)
		}
	}
	return res, nil
}

func DeleteDeployment(ip string) error {
	mu.Lock()
	defer mu.Unlock()
	return os.Remove(filepath.Join(dataDir, "deployments", safeIP(ip)+".json"))
}

func SaveAssetMeta(a Asset) error {
	mu.Lock()
	defer mu.Unlock()
	b, _ := json.MarshalIndent(a, "", "  ")
	return os.WriteFile(filepath.Join(dataDir, "assets", "_meta_"+a.ID+".json"), b, 0644)
}

func LoadAssets() ([]Asset, error) {
	mu.Lock()
	defer mu.Unlock()
	files, err := os.ReadDir(filepath.Join(dataDir, "assets"))
	if err != nil {
		return nil, err
	}
	var res []Asset
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			b, _ := os.ReadFile(filepath.Join(dataDir, "assets", f.Name()))
			var a Asset
			json.Unmarshal(b, &a)
			res = append(res, a)
		}
	}
	return res, nil
}

func DeleteAsset(id string) error {
	mu.Lock()
	defer mu.Unlock()
	
	// load meta to get filename
	b, err := os.ReadFile(filepath.Join(dataDir, "assets", "_meta_"+id+".json"))
	if err == nil {
		var a Asset
		json.Unmarshal(b, &a)
		os.Remove(filepath.Join(dataDir, "assets", id+"_"+a.Filename))
	}

	return os.Remove(filepath.Join(dataDir, "assets", "_meta_"+id+".json"))
}

func TrackClient(ip string) error {
	mu.Lock()
	defer mu.Unlock()
	
	clientsFile := filepath.Join(dataDir, "clients.json")
	b, err := os.ReadFile(clientsFile)
	var clients []Client
	if err == nil {
		json.Unmarshal(b, &clients)
	}
	
	now := time.Now().Format(time.RFC3339)
	found := false
	for i, c := range clients {
		if c.IP == ip {
			clients[i].LastSeen = now
			found = true
			break
		}
	}
	if !found {
		clients = append(clients, Client{IP: ip, LastSeen: now})
	}
	
	b, _ = json.MarshalIndent(clients, "", "  ")
	return os.WriteFile(clientsFile, b, 0644)
}

func GetRecentClients(days int) ([]Client, error) {
	mu.Lock()
	defer mu.Unlock()
	
	clientsFile := filepath.Join(dataDir, "clients.json")
	b, err := os.ReadFile(clientsFile)
	var clients []Client
	if err != nil {
		return clients, nil
	}
	json.Unmarshal(b, &clients)
	
	var recent []Client
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	for _, c := range clients {
		t, err := time.Parse(time.RFC3339, c.LastSeen)
		if err == nil && t.After(cutoff) {
			recent = append(recent, c)
		}
	}
	return recent, nil
}
