package file

import (
	"os"
	"strings"
)

// BOTPATH BOT当前路径
var path, _ = os.Getwd()
var BOTPATH = strings.ReplaceAll(path, "\\", "/")

// IsExist 文件/路径存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// IsNotExist 文件/路径不存在
func IsNotExist(path string) bool {
	_, err := os.Stat(path)
	return err != nil && os.IsNotExist(err)
}

// Size 获取文件大小
func Size(path string) (n int64) {
	stat, err := os.Stat(path)
	if err != nil {
		return
	}
	n = stat.Size()
	return
}

func CreateIfNotExist(paths ...string) {
	for _, path := range paths {
		if IsNotExist(path) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				panic(err)
			}
		}
	}
}
