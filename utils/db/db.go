package database

import (
	_ "MiniBot/utils/log"
	"MiniBot/utils/path"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
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

type Logger struct {
	logger zerolog.Logger
}

func (l Logger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

func (l Logger) Error(ctx context.Context, msg string, opts ...interface{}) {
	l.logger.Error().Msg(fmt.Sprintf(msg, opts...))
}

func (l Logger) Warn(ctx context.Context, msg string, opts ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprintf(msg, opts...))
}

func (l Logger) Info(ctx context.Context, msg string, opts ...interface{}) {
	l.logger.Info().Msg(fmt.Sprintf(msg, opts...))
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// if logger.LogLevel(zl.logger.GetLevel()) <= 0 {
	// 	return
	// }

	elapsed := time.Since(begin) // 执行时间
	sql, rows := fc()            // 获取 SQL 语句和影响的行数

	switch {
	case err != nil:
		l.logger.Error().
			Err(err).
			Str("sql", sql).
			Int64("rows", rows).
			Str("elapsed", fmt.Sprintf("%v", elapsed)).
			Msg("SQL execution error")
	case elapsed > time.Second: // 如果执行时间超过 1 秒，记录为 Warn 日志
		l.logger.Warn().
			Str("sql", sql).
			Int64("rows", rows).
			Str("elapsed", fmt.Sprintf("%v", elapsed)).
			Msg("Slow SQL query")

	default:
		l.logger.Info().
			Str("sql", sql).
			Int64("rows", rows).
			Str("elapsed", fmt.Sprintf("%v", elapsed)).
			Msg("SQL query executed")
	}

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
	gormLogger := &Logger{logger: log.Logger}
	gormLogger.LogMode(logger.Info)
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
