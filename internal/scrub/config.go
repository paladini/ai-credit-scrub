package scrub

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is stored in .ai-credit-scrub.yml. Custom rules must be explicitly
// reviewed after a scan before they can participate in automatic rewriting.
type Config struct {
	Version  int      `yaml:"version"`
	Reviewed bool     `yaml:"reviewed"`
	Literals []string `yaml:"literals"`
	Regex    []string `yaml:"regex"`
	Exclude  []string `yaml:"exclude"`
}

func Load(path string) (Config, error) {
	if path == "" {
		return Config{Version: 1}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("config version must be 1")
	}
	if (len(c.Literals) > 0 || len(c.Regex) > 0) && !c.Reviewed {
		return fmt.Errorf("custom rules require reviewed: true; run scan first and review every match")
	}
	return nil
}
