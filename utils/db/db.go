package database

import (
	log_utils "MiniBot/utils/log"
	"MiniBot/utils/path"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var utilsName = "DB"

type DBInfo struct {
	Dsn  string `yaml:"dsn"`
	Type string `yaml:"type"`
	Db   *gorm.DB
}

type Config struct {
	DbType map[string]bool   `yaml:"db_type"`
	DbMap  map[string]DBInfo `yaml:"db"`
}

func NewDbConfig() *Config {
	config := &Config{}
	yamlPath := filepath.Join(path.ConfPath, "db.yaml")
	if os.Getenv("ENVIRONMENT") == "dev" {
		yamlPath = filepath.Join(path.ConfPath, "db_dev.yaml")
	}
	yamlFile, err := os.ReadFile(yamlPath)
	if err != nil {
		log.Error().Str("name", utilsName).Err(err).Msg("")
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Error().Str("name", utilsName).Err(err).Msg("")
	}
	gormLogger := &log_utils.GormLogger{}
	for k, v := range config.DbMap {
		//	ok2代表是否有该数据库，ok1代表是否开启该数据库，实际上不用的话，可以直接将数据库删除
		if ok1, ok2 := config.DbType[v.Type]; ok2 && ok1 {
			switch {
			case v.Type == "pgsql":
				v.Db, err = gorm.Open(postgres.Open(v.Dsn), &gorm.Config{
					Logger: gormLogger.LogMode(logger.Info),
				})

				if err != nil {
					log.Error().Str("name", utilsName).Err(err).Msgf("Error open database %v", k)
					continue
				}
				log.Info().Str("name", utilsName).Msgf("success open database %v", k)
			case v.Type == "mysql":
				v.Db, err = gorm.Open(mysql.Open(v.Dsn), &gorm.Config{})
				if err != nil {
					log.Error().Str("name", utilsName).Err(err).Msgf("Error open database %v", k)
					continue
				}
				log.Info().Str("name", utilsName).Msgf("success open database %v", k)
			}
			config.DbMap[k] = v
		}
	}
	return config
}

var DbConfig = NewDbConfig()

func (config *Config) GetDb(name string) *gorm.DB {
	if dbInfo, ok := config.DbMap[name]; ok {
		return dbInfo.Db
	}
	return nil
}

func init() {

}
