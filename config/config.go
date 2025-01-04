package config

import (
	"MiniBot/utils/path"
	zero "ZeroBot"
	"os"

	"ZeroBot/driver"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var Config MiniConfig

type ConnConfig struct {
	URL         string `yaml:"url"`
	AccessToken string `yaml:"access_token"`
}

// 对应config.json
type MiniConfig struct {
	Z   zero.Config  `json:"zero" yaml:"zero" mapstructure:"zero"`
	WS  []ConnConfig `json:"ws" yaml:"ws"`
	WSS []ConnConfig `json:"wss" yaml:"wss"`
	// C   []*driver.WSClient `json:"-" yaml:"-"`
	// S   []*driver.WSServer `json:"-" yaml:"-"`
}

func ConfigInit() {
	// Read YAML file
	configName := "config"
	if os.Getenv("ENVIRONMENT") == "dev" {
		configName = "config_dev"
		log.Info().Msg("目前处于开发环境，请注意主目录下所有配置文件的内容和路径是否正确")
	}

	viper.SetConfigName(configName)    // 配置文件名称(无后缀)
	viper.AddConfigPath(path.ConfPath) // 配置文件路径
	viper.SetConfigType("yaml")        // 配置文件后缀, 也可以是 json, toml, yaml, yml 等, 不设置则自动识别

	// 读取配置文件, 如果出错则退出
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Msgf("Error reading config file, %s\n", err)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatal().Msgf("Error reading config file, %s\n", err)
	}

	for _, client := range Config.WS {
		Config.Z.Driver = append(Config.Z.Driver, driver.NewWebSocketClient(client.URL, client.AccessToken))
	}
	for _, server := range Config.WSS {
		Config.Z.Driver = append(Config.Z.Driver, driver.NewWebSocketServer(5, server.URL, server.AccessToken))
	}
}
