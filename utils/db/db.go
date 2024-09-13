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

type zerologLogger struct {
	logger zerolog.Logger
}

func (zl *zerologLogger) LogMode(level logger.LogLevel) logger.Interface {
	zl.logger.WithLevel(zerolog.Level(level))
	return zl
}

// Info 实现 GORM 的 Info 日志
func (zl *zerologLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if logger.LogLevel(zl.logger.GetLevel()) >= logger.Info {
		zl.logger.Info().Msgf(msg, data...)
	}
}

// Warn 实现 GORM 的 Warn 日志
func (zl *zerologLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if logger.LogLevel(zl.logger.GetLevel()) >= logger.Warn {
		zl.logger.Warn().Msgf(msg, data...)
	}
}

// Error 实现 GORM 的 Error 日志
func (zl *zerologLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if logger.LogLevel(zl.logger.GetLevel()) >= logger.Error {
		zl.logger.Error().Msgf(msg, data...)
	}
}

// Trace 实现 GORM 的 Trace 日志（包括 SQL 语句、执行时间、错误等）
func (zl *zerologLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if logger.LogLevel(zl.logger.GetLevel()) <= 0 {
		return
	}

	elapsed := time.Since(begin) // 执行时间
	sql, rows := fc()            // 获取 SQL 语句和影响的行数

	switch {
	case err != nil:
		if logger.LogLevel(zl.logger.GetLevel()) >= logger.Error {
			zl.logger.Error().
				Err(err).
				Str("sql", sql).
				Int64("rows", rows).
				Str("elapsed", fmt.Sprintf("%v", elapsed)).
				Msg("SQL execution error")
		}
	case elapsed > time.Second: // 如果执行时间超过 1 秒，记录为 Warn 日志
		if logger.LogLevel(zl.logger.GetLevel()) >= logger.Warn {
			zl.logger.Warn().
				Str("sql", sql).
				Int64("rows", rows).
				Str("elapsed", fmt.Sprintf("%v", elapsed)).
				Msg("Slow SQL query")
		}
	default:
		if logger.LogLevel(zl.logger.GetLevel()) >= logger.Info {
			zl.logger.Info().
				Str("sql", sql).
				Int64("rows", rows).
				Str("elapsed", fmt.Sprintf("%v", elapsed)).
				Msg("SQL query executed")
		}
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
	gormLogger := &zerologLogger{logger: log.Logger}
	for k, v := range config.DbMap {
		//	ok2代表是否有该数据库，ok1代表是否开启该数据库，实际上不用的话，可以直接将数据库删除
		if ok1, ok2 := config.DbType[v.Type]; ok2 && ok1 {
			switch {
			case v.Type == "pgsql":
				v.Db, err = gorm.Open(postgres.Open(v.Dsn), &gorm.Config{Logger: gormLogger})
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
