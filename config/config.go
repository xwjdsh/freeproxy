package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App       *AppConfig       `yaml:"app"`
	Parser    *ParserConfig    `yaml:"parser"`
	Validator *ValidatorConfig `yaml:"validator"`
	Storage   *StorageConfig   `yaml:"storage"`
	Exporter  *ExporterConfig  `yaml:"exporter"`
	Log       *LogConfig       `yaml:"log"`
}

func (c *Config) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

type AppConfig struct {
	Worker int `yaml:"worker"`
}

type LogConfig struct {
	Level zapcore.Level `yaml:"level"`
}

type ParserExecutor struct {
	Name string `yaml:"name"`
}

type ParserConfig struct {
	Executors []*ParserExecutor `yaml:"executors"`
}

type ValidatorConfig struct {
	TestNetworkURL        string        `yaml:"test_network_url"`
	TestURL               string        `yaml:"test_url"`
	TestURLCount          int           `yaml:"test_url_count"`
	TestURLTimeout        time.Duration `yaml:"test_url_timeout"`
	GetCountryInfoTimeout time.Duration `yaml:"get_country_info_timeout"`
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
		App: &AppConfig{
			Worker: 100,
		},
		Parser: &ParserConfig{
			Executors: []*ParserExecutor{
				{Name: "cfmem"},
				{Name: "freefq"},
				{Name: "feedburner"},
			},
		},
		Validator: &ValidatorConfig{
			TestNetworkURL:        "https://www.baidu.com",
			TestURL:               "http://www.gstatic.com/generate_204",
			TestURLCount:          3,
			TestURLTimeout:        5 * time.Second,
			GetCountryInfoTimeout: 5 * time.Second,
		},
		Storage: &StorageConfig{
			Driver: "sqlite",
			DSN:    fmt.Sprintf("%s/.config/freeproxy/freeproxy.db", homeDir),
		},
		Exporter: &ExporterConfig{},
		Log: &LogConfig{
			Level: zapcore.InfoLevel,
		},
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
