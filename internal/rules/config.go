package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/client9/tojson"
)

// Config holds project-level plint configuration from .plint/config.yaml.
type Config struct {
	Pipeline []string `json:"pipeline"`
}

// LoadConfig reads .plint/config.yaml (or .yml) from plintDir.
// Returns (nil, nil) if no config file exists.
func LoadConfig(plintDir string) (*Config, error) {
	var path string
	for _, name := range []string{"config.yaml", "config.yml"} {
		p := filepath.Join(plintDir, name)
		if _, err := os.Stat(p); err == nil {
			path = p
			break
		}
	}
	if path == "" {
		return nil, nil
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := tojson.FromYAML(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(jsonBytes, &cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &cfg, nil
}
