package config

import (
	"MiniBot/utils/path"
	zero "ZeroBot"
	"os"
	"path/filepath"

	"ZeroBot/driver"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var Config MiniConfig

type ConnConfig struct {
	URL         string `yaml:"url"`
	AccessToken string `yaml:"access_token"`
}

// 对应config.json
type MiniConfig struct {
	Z   zero.Config  `json:"zero" yaml:"zero"`
	WS  []ConnConfig `json:"ws" yaml:"ws"`
	WSS []ConnConfig `json:"wss" yaml:"wss"`
	// C   []*driver.WSClient `json:"-" yaml:"-"`
	// S   []*driver.WSServer `json:"-" yaml:"-"`
}

func ConfigInit() {
	// Read YAML file
	configPath := filepath.Join(path.ConfPath, "config.yaml")
	if os.Getenv("ENVIRONMENT") == "dev" {
		configPath = filepath.Join(path.ConfPath, "config_dev.yaml")
		logrus.Infoln("目前处于开发环境，请注意主目录下所有配置文件的内容和路径是否正确")
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}

	// Unmarshal YAML data
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		logrus.Fatalf("error unmarshalling config file: %v", err)
	}
	for _, client := range Config.WS {
		Config.Z.Driver = append(Config.Z.Driver, driver.NewWebSocketClient(client.URL, client.AccessToken))
	}
	for _, server := range Config.WSS {
		Config.Z.Driver = append(Config.Z.Driver, driver.NewWebSocketServer(5, server.URL, server.AccessToken))
	}
}
