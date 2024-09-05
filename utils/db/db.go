package database

import (
	"MiniBot/utils/path"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
		logrus.Errorf("Error reading YAML file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		logrus.Errorf("Error parsing YAML: %v", err)
	}
	for k, v := range config.DbMap {
		//	ok2代表是否有该数据库，ok1代表是否开启该数据库，实际上不用的话，可以直接将数据库删除
		if ok1, ok2 := config.DbType[v.Type]; ok2 && ok1 {
			switch {
			case v.Type == "pgsql":
				v.Db, err = gorm.Open(postgres.Open(v.Dsn), &gorm.Config{})
				if err != nil {
					logrus.Errorf("Error open database %v", k)
					continue
				}
				logrus.Infof("success open database %v", k)
			case v.Type == "mysql":
				v.Db, err = gorm.Open(mysql.Open(v.Dsn), &gorm.Config{})
				if err != nil {
					logrus.Errorf("Error open database %v", k)
					continue
				}
				logrus.Infof("success open database %v", k)
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
