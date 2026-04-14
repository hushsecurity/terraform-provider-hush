package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const defaultFixturesPath = "testdata/mock_api_fixtures.json"

// Fixtures holds parsed mock API fixture data.
type Fixtures struct {
	Meta           map[string]any            `json:"_meta"`
	ComputedFields map[string]map[string]any `json:"computed_fields"`
	Endpoints      map[string]map[string]any `json:"endpoints"`
}

// LoadFixtures loads mock API fixtures from MOCK_FIXTURES_PATH env or default location.
func LoadFixtures() (*Fixtures, error) {
	path := os.Getenv("MOCK_FIXTURES_PATH")
	if path == "" {
		root, err := findRepoRoot()
		if err != nil {
			return nil, fmt.Errorf("fixtures not found: run 'make fetch-mock-fixtures' or set MOCK_FIXTURES_PATH")
		}
		path = filepath.Join(root, defaultFixturesPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("fixtures not found: run 'make fetch-mock-fixtures' or set MOCK_FIXTURES_PATH")
		}
		return nil, fmt.Errorf("reading fixtures: %w", err)
	}

	var f Fixtures
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing fixtures: %w", err)
	}
	return &f, nil
}

// GetComputedFields returns computed fields for a resource type with placeholders resolved.
func (f *Fixtures) GetComputedFields(resourceType, uuid string) map[string]any {
	fields, ok := f.ComputedFields[resourceType]
	if !ok {
		return nil
	}
	result := make(map[string]any, len(fields))
	ts := time.Now().UTC().Format(time.RFC3339)
	for k, v := range fields {
		if s, ok := v.(string); ok {
			s = strings.ReplaceAll(s, "{uuid}", uuid)
			s = strings.ReplaceAll(s, "{timestamp}", ts)
			result[k] = s
		} else {
			result[k] = v
		}
	}
	return result
}

// findRepoRoot walks up from caller directory to find go.mod.
func findRepoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to determine caller path")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
