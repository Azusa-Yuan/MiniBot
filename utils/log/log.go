package log

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
)

func init() {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

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
}

/*
log文件设置 format:log/level2006-01-02.log
*/
func writer(logPath string, level string, save int) *rotatelogs.RotateLogs {
	logFullPath := filepath.Join(logPath, level)
	// var cstSh, _ = time.LoadLocation("Asia/Shanghai") //上海
	// fileSuffix := time.Now().In(cstSh).Format("2006-01-02") + ".log"

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
