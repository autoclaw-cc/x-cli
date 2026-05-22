package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"scholar-cli/paper"
)

type Store struct {
	Dir    string
	Papers []paper.Paper
}

func DefaultDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "scholar-cli", "default")
}

func Open(dir string) (*Store, error) {
	if dir == "" {
		dir = DefaultDir()
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	s := &Store{Dir: dir}
	papersFile := filepath.Join(dir, "papers.json")
	data, err := os.ReadFile(papersFile)
	if err == nil {
		json.Unmarshal(data, &s.Papers)
	}
	return s, nil
}

func (s *Store) Save() error {
	data, err := json.MarshalIndent(s.Papers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.Dir, "papers.json"), data, 0644)
}

func (s *Store) AddPapers(papers []paper.Paper) int {
	combined := append(s.Papers, papers...)
	deduped := paper.Deduplicate(combined)
	added := len(deduped) - len(s.Papers)
	s.Papers = deduped
	return added
}

func (s *Store) Count() int {
	return len(s.Papers)
}
