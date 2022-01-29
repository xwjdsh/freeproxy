package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ValidatorCount int              `yaml:"validator_count"`
	StorageCount   int              `yaml:"storage_count"`
	Parser         *ParserConfig    `yaml:"parser"`
	Validator      *ValidatorConfig `yaml:"validator"`
	Storage        *StorageConfig   `yaml:"storage"`
	Exporter       *ExporterConfig  `yaml:"exporter"`
}

func (c *Config) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

type ParserExecutor struct {
	Name string `yaml:"name"`
}

type ParserConfig struct {
	Executors []*ParserExecutor `yaml:"executors"`
}

type ValidatorConfig struct {
	TestURL string        `yaml:"test_url"`
	Timeout time.Duration `yaml:"timeout"`
}

type StorageConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type ExporterConfig struct {
	TemplateFilePath string `yaml:"template_file_path"`
	OutputFilePath   string `yaml:"output_file_path"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		StorageCount:   10,
		ValidatorCount: 10,
		Parser: &ParserConfig{
			Executors: []*ParserExecutor{
				{Name: "cfmem"},
			},
		},
		Validator: &ValidatorConfig{
			TestURL: "http://www.gstatic.com/generate_204",
			Timeout: 5 * time.Second,
		},
		Storage: &StorageConfig{
			Driver: "sqlite",
			DSN:    fmt.Sprintf("%s/.config/freeproxy/freeproxy.db", homeDir),
		},
		Exporter: &ExporterConfig{},
	}
}

func Init(fp string) (*Config, error) {
	cfg := DefaultConfig()
	data, err := ioutil.ReadFile(fp)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}