package registry

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var notebookPath = regexp.MustCompile(`^/notebook/([0-9a-fA-F-]{36})/?$`)

type Notebook struct {
	ID           string    `json:"id"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	AuthorizedAt time.Time `json:"authorized_at"`
}

type Registry struct {
	path      string
	Notebooks []Notebook `json:"notebooks"`
}

func Load(path string) (*Registry, error) {
	r := &Registry{path: path, Notebooks: []Notebook{}}
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return r, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read registry: %w", err)
	}
	if err := json.Unmarshal(body, r); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}
	r.path = path
	return r, nil
}

func (r *Registry) Authorize(rawURL, title string) (*Notebook, error) {
	u, err := parseNotebookURL(rawURL)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	id := notebookPath.FindStringSubmatch(u.Path)[1]
	for i := range r.Notebooks {
		if r.Notebooks[i].ID == id {
			r.Notebooks[i].URL = u.String()
			r.Notebooks[i].Title = title
			r.Notebooks[i].AuthorizedAt = time.Now().UTC()
			return &r.Notebooks[i], nil
		}
	}
	r.Notebooks = append(r.Notebooks, Notebook{
		ID: id, URL: u.String(), Title: title, AuthorizedAt: time.Now().UTC(),
	})
	return &r.Notebooks[len(r.Notebooks)-1], nil
}

func (r *Registry) RequireOwned(id string) (*Notebook, error) {
	for i := range r.Notebooks {
		if r.Notebooks[i].ID == id {
			return &r.Notebooks[i], nil
		}
	}
	return nil, fmt.Errorf("notebook_not_owned: notebook %q is not in the local ownership registry", id)
}

func (r *Registry) Save() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return fmt.Errorf("create registry directory: %w", err)
	}
	body, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("encode registry: %w", err)
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, append(body, '\n'), 0o600); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}
	if err := os.Rename(tmp, r.path); err != nil {
		return fmt.Errorf("replace registry: %w", err)
	}
	return nil
}

func DefaultPath() string {
	if base := os.Getenv("LOCALAPPDATA"); base != "" {
		return filepath.Join(base, "notebooklm-cli", "owned-notebooks.json")
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".", ".notebooklm-cli", "owned-notebooks.json")
	}
	return filepath.Join(base, "notebooklm-cli", "owned-notebooks.json")
}

func parseNotebookURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid notebook URL: %w", err)
	}
	if u.Scheme != "https" || u.Host != "notebooklm.google.com" || !notebookPath.MatchString(u.Path) {
		return nil, fmt.Errorf("invalid NotebookLM notebook URL")
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u, nil
}
