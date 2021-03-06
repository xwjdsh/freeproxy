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
	Log       *LogConfig       `yaml:"log"`
}

func (c *Config) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

type AppConfig struct {
	Fetch   *AppFetchConfig   `yaml:"fetch"`
	Tidy    *AppTidyConfig    `yaml:"tidy"`
	Export  *AppExportConfig  `yaml:"export"`
	Proxy   *AppProxyConfig   `yaml:"proxy"`
	Summary *AppSummaryConfig `yaml:"summary"`
}

type AppFetchConfig struct {
	Worker int `yaml:"worker"`
}

type AppTidyConfig struct {
	Worker int `yaml:"worker"`
}

type AppSummaryConfig struct {
	TemplateFilePath string `yaml:"template_file_path"`
}

type AppProxyConfig struct {
	BindAddress          string `yaml:"bind_address"`
	Port                 int    `yaml:"port"`
	SwitchServer         string `yaml:"switch_server"`
	Verbose              bool   `yaml:"verbose"`
	ProxyCountryCodes    string `yaml:"proxy_country_codes"`
	ProxyNotCountryCodes string `yaml:"proxy_not_country_codes"`
	ProxyID              uint   `yaml:"proxy_id"`
}

type AppExportConfig struct {
	TemplateFilePath     string `yaml:"template_file_path"`
	OutputFilePath       string `yaml:"output_file_path"`
	ProxyCount           int    `yaml:"proxy_count"`
	ProxyCountryCodes    string `yaml:"proxy_country_codes"`
	ProxyNotCountryCodes string `yaml:"proxy_not_country_codes"`
	ProxyID              uint   `yaml:"proxy_id"`
}

type LogConfig struct {
	Level zapcore.Level `yaml:"level"`
}

type ParserExecutor struct {
	Name    string        `yaml:"name"`
	Enable  bool          `yaml:"enable"`
	Timeout time.Duration `yaml:"deadline"`
	FileURL string        `yaml:"file_url"`
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

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	c := &Config{
		App: &AppConfig{
			Fetch:  &AppFetchConfig{Worker: 300},
			Tidy:   &AppTidyConfig{Worker: 300},
			Export: &AppExportConfig{ProxyCount: 100},
			Proxy: &AppProxyConfig{
				BindAddress:  "127.0.0.1",
				Port:         10000,
				SwitchServer: "127.0.0.1:9999",
			},
			Summary: &AppSummaryConfig{},
		},
		Parser: &ParserConfig{
			Executors: []*ParserExecutor{
				{Name: "cfmem"},
				{Name: "freefq_ss"},
				{Name: "freefq_ssr"},
				{Name: "freefq_v2ray"},
				{Name: "feedburner"},
				{Name: "freefq/free/v2", FileURL: "https://raw.githubusercontent.com/freefq/free/master/v2"},
				{Name: "aiboboxx/v2rayfree", FileURL: "https://raw.githubusercontent.com/aiboboxx/v2rayfree/main/v2"},
				{Name: "learnhard-cn/free_proxy_ss/ss", FileURL: "https://raw.githubusercontent.com/learnhard-cn/free_proxy_ss/main/ss/sssub"},
				{Name: "learnhard-cn/free_proxy_ss/ssr", FileURL: "https://raw.githubusercontent.com/learnhard-cn/free_proxy_ss/main/ssr/ssrsub"},
				{Name: "learnhard-cn/free_proxy_ss/v2ray", FileURL: "https://raw.githubusercontent.com/learnhard-cn/free_proxy_ss/main/v2ray/v2raysub"},
				{Name: "learnhard-cn/free_proxy_ss/free", FileURL: "https://raw.githubusercontent.com/learnhard-cn/free_proxy_ss/main/free"},
				{Name: "chfchf0306/jeidian4.18", FileURL: "https://raw.githubusercontent.com/chfchf0306/jeidian4.18/main/4.18"},
				{Name: "xiyaowong/freeFQ", FileURL: "https://raw.githubusercontent.com/xiyaowong/freeFQ/main/v2ray"},
				{Name: "vpei/Free-Node-Merge", FileURL: "https://raw.githubusercontent.com/vpei/Free-Node-Merge/main/out/node.txt"},
				{Name: "colatiger/v2ray-nodes/proxy", FileURL: "https://raw.githubusercontent.com/colatiger/v2ray-nodes/master/proxy.md"},
				{Name: "colatiger/v2ray-nodes/ss", FileURL: "https://raw.githubusercontent.com/colatiger/v2ray-nodes/master/ss.md"},
				{Name: "colatiger/v2ray-nodes/vmess", FileURL: "https://raw.githubusercontent.com/colatiger/v2ray-nodes/master/vmess.md"},
				{Name: "ssrsub/ssr/V2Ray", FileURL: "https://raw.githubusercontent.com/ssrsub/ssr/master/V2Ray"},
				{Name: "ssrsub/ssr/V2Ray", FileURL: "https://raw.githubusercontent.com/ssrsub/ssr/master/V2Ray"},
				{Name: "ssrsub/ssr/ss-sub", FileURL: "https://raw.githubusercontent.com/ssrsub/ssr/master/ss-sub"},
				{Name: "ssrsub/ssr/ssrsub", FileURL: "https://raw.githubusercontent.com/ssrsub/ssr/master/ssrsub"},
				{Name: "ssrsub/ssr/trojan", FileURL: "https://raw.githubusercontent.com/ssrsub/ssr/master/trojan"},
				{Name: "Leon406/SubCrawler", FileURL: "https://raw.githubusercontent.com/Leon406/SubCrawler/main/sub/share/all"},
				{Name: "wrfree/free/ssr", FileURL: "https://raw.githubusercontent.com/wrfree/free/main/ssr"},
				{Name: "wrfree/free/v2", FileURL: "https://raw.githubusercontent.com/wrfree/free/main/v2"},
				{Name: "ThekingMX1998/free-v2ray-code", FileURL: "https://raw.githubusercontent.com/GreenFishStudio/GreenFish/master/Subscription/GreenFishYYDS"},
			},
		},
		Validator: &ValidatorConfig{
			TestNetworkURL:        "https://www.baidu.com",
			TestURL:               "http://www.gstatic.com/generate_204",
			TestURLCount:          3,
			TestURLTimeout:        10 * time.Second,
			GetCountryInfoTimeout: 5 * time.Second,
		},
		Storage: &StorageConfig{
			Driver: "sqlite",
			DSN:    fmt.Sprintf("%s/.config/freeproxy/freeproxy.db", homeDir),
		},
		Log: &LogConfig{
			Level: zapcore.InfoLevel,
		},
	}

	for _, e := range c.Parser.Executors {
		e.Timeout = 10 * time.Second
		e.Enable = true
	}

	return c
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
