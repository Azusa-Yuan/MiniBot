package log_utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
)

var logSourceDir string
var LogLevel string

func init() {
	LogLevel = os.Getenv("LogLevel")
	if LogLevel == "debug" {
		log.Info().Msg("目前处于debug等级，请注意打印日志的等级")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// 创建日志文件夹
	logFolder := "./log"
	err := os.MkdirAll(logFolder, 0755)
	if err != nil {
		log.Fatal().Msg("Error creating directory")
	}

	// 记录日志到文件
	log.Logger = log.Output(zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		writer("./log", "MiniBot", 10),
	))
	log.Logger = log.With().Caller().Logger()

	_, file, _, _ := runtime.Caller(0)
	// 通过当前文件的路径获取 GORM 源代码目录
	logSourceDir = sourceDir(file)
}

/*
日志切割设置 log文件设置 format:log/level2006-01-02.log
*/
func writer(logPath string, level string, save int) *rotatelogs.RotateLogs {
	logFullPath := filepath.Join(logPath, level)

	logier, err := rotatelogs.New(
		logFullPath+"-"+"%Y%m%d"+".log",
		rotatelogs.WithLinkName(logFullPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithRotationCount(save),        // 文件最大保存份数 负数到0说明保存无限
		rotatelogs.WithRotationTime(time.Hour*24), // 日志切割时间间隔
	)

	if err != nil {
		panic(err)
	}
	return logier
}

func sourceDir(file string) string {
	dir := filepath.Dir(file) // 获取当前文件所在目录
	dir = filepath.Dir(dir)   // 再获取上一级目录

	s := filepath.Dir(dir) // 获取上上级目录
	// 检查上上级目录是否为 "gorm.io"
	if filepath.Base(s) != "gorm.io" {
		s = dir // 如果不是，则使用上一级目录
	}
	return filepath.ToSlash(s) + "/" // 返回目录路径并将反斜杠转换为正斜杠
}

func FileWithLineNum() string {
	pcs := [13]uintptr{}
	// the third caller usually from gorm internal
	len := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:len])
	for i := 0; i < len; i++ {
		// second return value is "more", not "ok"
		frame, _ := frames.Next()
		if (!strings.HasPrefix(frame.File, logSourceDir) && !strings.Contains(frame.File, "gorm.io") ||
			strings.HasSuffix(frame.File, "_test.go")) && !strings.HasSuffix(frame.File, ".gen.go") {
			return string(strconv.AppendInt(append([]byte(frame.File), ':'), int64(frame.Line), 10))
		}
	}

	return ""
}

type GormLogger struct {
	Logger   zerolog.Logger
	LogLevel logger.LogLevel
}

func (l GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

func (l GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel > logger.Error {
		l.Logger.Error().Str("sql_caller", FileWithLineNum()).Msg(fmt.Sprintf(msg, data...))
	}
}

func (l GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel > logger.Warn {
		l.Logger.Warn().Str("sql_caller", FileWithLineNum()).Msg(fmt.Sprintf(msg, data...))
	}
}

func (l GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel > logger.Info {
		l.Logger.Info().Str("sql_caller", FileWithLineNum()).Msg(fmt.Sprintf(msg, data...))
	}
}

func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin) // 执行时间
	sql, rows := fc()            // 获取 SQL 语句和影响的行数

	switch {
	// todo  不知道具体是哪一行代码出错
	case err != nil:
		l.Logger.Error().Str("sql_caller", FileWithLineNum()).
			Err(err).
			Str("sql", sql).
			Int64("rows", rows).
			Str("elapsed", fmt.Sprintf("%v", elapsed)).
			Msg("SQL execution error")
	case elapsed > 200*time.Millisecond: // 如果执行时间超过 1 秒，记录为 Warn 日志
		l.Logger.Warn().Str("sql_caller", FileWithLineNum()).
			Str("sql", sql).
			Int64("rows", rows).
			Str("elapsed", fmt.Sprintf("%v", elapsed)).
			Msg("Slow SQL query")

	default:
		l.Logger.Info().Str("sql_caller", FileWithLineNum()).
			Str("sql", sql).
			Int64("rows", rows).
			Str("elapsed", fmt.Sprintf("%v", elapsed)).
			Msg("SQL query executed")
	}

}
