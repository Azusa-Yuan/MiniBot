package path

import (
	"MiniBot/utils/file"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// 公共路径
var (
	// 插件文件夹路径
	PluginPath = "/plugin"
	// 数据存储路径
	DataPath = "/data"
	// 公共图片存储路径
	ImgPath = "/img"
	// bot运行路径
	PWDPath = "/"
	// bot的配置文件夹
	ConfPath = "/config"
	kinds    = []string{"plugin", "service", "utils"}
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal("Error getting current directory:", err)
		return
	}
	// 查找最后一个 "plugin" 的起始位置
	lastIndex := strings.LastIndex(dir, "plugin")
	if lastIndex != -1 {
		dir = dir[0 : lastIndex-1]
		logrus.Infoln("[path] 根据当前工作路径认定为是测试环境,路径有所更改 请注意")
	}
	lastIndex = strings.LastIndex(dir, "service")
	if lastIndex != -1 {
		dir = dir[0 : lastIndex-1]
		logrus.Infoln("[path] 包根据当前工作路径认定为是测试环境,路径有所更改 请注意")
	}
	lastIndex = strings.LastIndex(dir, "utils")
	if lastIndex != -1 {
		dir = dir[0 : lastIndex-1]
		logrus.Infoln("[path] 包根据当前工作路径认定为是测试环境,路径有所更改 请注意")
	}
	PluginPath = dir + PluginPath
	DataPath = dir + DataPath
	ImgPath = dir + ImgPath
	PWDPath = dir + PWDPath
	ConfPath = dir + ConfPath

	file.CreateIfNotExist(PluginPath, DataPath, ImgPath)
}

func GetPluginDataPath() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to get caller")
	}
	name := runtime.FuncForPC(pc).Name()
	a := strings.LastIndex(name, "plugin")
	if a < 0 {
		panic("invalid package name: " + name)
	}
	name = name[a+7:]
	b := strings.Index(name, ".")
	if b < 0 {
		panic("invalid package name: " + name)
	}
	name = name[:b]
	path := filepath.Join(DataPath, name)
	file.CreateIfNotExist(path)
	return path
}

// path 包内部使用
func getDataPath(packagePath string, kind string) (string, error) {

	path := ""
	a := strings.LastIndex(packagePath, kind)
	if a < 0 {
		return path, fmt.Errorf("该包不属于%s", kind)
	}
	path = packagePath[a+len(kind):]
	b := strings.Index(path, ".")
	if b < 0 {
		return path, fmt.Errorf("invalid package name:%s", path)
	}
	path = path[:b]
	path = filepath.Join(DataPath, path)
	file.CreateIfNotExist(path)
	return path, nil
}

// 判断优先级  plugin，service，utils
func GetDataPath() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to get caller")
	}
	name := runtime.FuncForPC(pc).Name()
	var path string
	var err error
	for _, kind := range kinds {
		path, err = getDataPath(name, kind)
		if err == nil {
			logrus.Debugln("判断包属于", kind)
			break
		}
	}
	if err != nil {
		panic("invalid package name: " + name)
	}
	return path
}
