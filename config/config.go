package config

import (
	"os"

	common "github.com/opensourceways/image-scanning/common/config"
	"github.com/opensourceways/image-scanning/common/infrastructure/postgresql"
	"github.com/opensourceways/image-scanning/scanning/app"
	"github.com/opensourceways/image-scanning/scanning/domain"
	"github.com/opensourceways/image-scanning/utils"
)

// LoadConfig loads the configuration file from the specified path and deletes the file if needed
func LoadConfig(path string, cfg *Config, remove bool) error {
	if remove {
		defer os.Remove(path)
	}

	if err := utils.LoadFromYaml(path, cfg); err != nil {
		return err
	}

	common.SetDefault(cfg)

	return common.Validate(cfg)
}

type Config struct {
	Community   []domain.Community `json:"community"`
	TrivyRepo   app.TrivyRepo      `json:"trivy_repo"`
	Postgresql  postgresql.Config  `json:"postgresql"`
	Concurrency app.Concurrency    `json:"concurrency"`
}

// ConfigItems returns a slice of interface{} containing pointers to the configuration items.
func (cfg *Config) ConfigItems() []interface{} {
	return []interface{}{
		&cfg.Community,
		&cfg.TrivyRepo,
		&cfg.Concurrency,
	}
}

// SetDefault sets default values for the Config struct.
func (cfg *Config) SetDefault() {
}

// Validate validates the configuration.
func (cfg *Config) Validate() error {
	return common.CheckConfig(cfg, "")
}
