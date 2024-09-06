package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/sirupsen/logrus"
)

func init() {
	logLevel := logrus.DebugLevel

	logrus.SetLevel(logLevel)

	// 设置输出格式为自定义的TextFormatter
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,                     // 禁用颜色
		FullTimestamp:   true,                      // 显示完整时间戳
		TimestampFormat: "2006-01-02T15:04:05.000", // 时间戳格式，精确到毫秒
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf(" %s:%d", filepath.Base(f.File), f.Line)
		},
	})

	logFolder := "./log"
	err := os.MkdirAll(logFolder, 0755)
	if err != nil {
		logrus.Fatalf("Error creating directory")
	}

	levelList := []logrus.Level{}
	for i := 0; i <= int(logLevel); i++ {
		levelList = append(levelList, logrus.Level(i))
	}
	criticalHook := FileHook{
		writer: writer("./log", "critical", 10),
		levels: levelList,
	}
	logrus.AddHook(&criticalHook)

	errHook := FileHook{
		writer: writer("./log", "error", 10),
		levels: []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel},
	}

	logrus.AddHook(&errHook)

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

type FileHook struct {
	writer io.Writer
	levels []logrus.Level
}

func (hook *FileHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *FileHook) Fire(entry *logrus.Entry) error {
	// 创建一个TextFormatter（不带颜色）来输出到文件
	formatter := &logrus.TextFormatter{
		DisableColors:   true, // 禁用颜色
		FullTimestamp:   true, // 显示完整时间戳
		TimestampFormat: "2006-01-02T15:04:05.000",
	}

	// 格式化日志条目
	line, err := formatter.Format(entry)
	if err != nil {
		return err
	}

	// 写入文件
	_, err = hook.writer.Write(line)
	if err != nil {
		return err
	}

	return nil
}
