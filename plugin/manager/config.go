package manager

import (
	"MiniBot/utils/path"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	BlockStranger  bool `yaml:"block_stranger"`
	BlockMalicious bool `yaml:"block_malicious"`
}

// manager config
var MC Config

func init() {
	configPath := filepath.Join(path.GetDataPath(), "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Error().Str("name", pluginName).Msg("")
	}
	err = yaml.Unmarshal(data, &MC)
	if err != nil {
		log.Error().Str("name", pluginName).Msg("")
	}
}
