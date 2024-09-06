package cache

import (
	"MiniBot/utils"
	"MiniBot/utils/net_tools"
	"MiniBot/utils/path"
	"MiniBot/utils/text"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type rdb struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Config struct {
	Rdb rdb `yaml:"redis"`
}

var (
	redisClient     *redis.Client
	defalutFontPath = text.MaokenFontFile
	// 以下不需要填，由init完成
)

func init() {
	config := &Config{}
	yamlPath := filepath.Join(path.ConfPath, "cache.yaml")
	if os.Getenv("ENVIRONMENT") == "dev" {
		yamlPath = filepath.Join(path.ConfPath, "cache_dev.yaml")
	}

	yamlFile, err := os.ReadFile(yamlPath)
	if err != nil {
		logrus.Fatalf("Error reading YAML file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		logrus.Fatalf("Error parsing YAML: %v", err)
	}

	redisClientTmp := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Rdb.Host, config.Rdb.Port),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	// Ping Redis server
	_, err = redisClientTmp.Ping(ctx).Result()
	if err != nil {
		logrus.Fatalf("Failed to connect to Redis: %v", err)
	}
	logrus.Infoln("connect redis success!")
	redisClient = redisClientTmp
}

func GetAvatar(uid int64) ([]byte, error) {
	uidStr := strconv.FormatInt(uid, 10)
	ctx := context.TODO()
	res, err := redisClient.Get(ctx, uidStr+"avatar").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if res != "" {
		return utils.StringToBytes(res), nil
	}
	url := "http://q4.qlogo.cn/g?b=qq&nk=" + uidStr + "&s=640"
	respBytes, err := net_tools.Download(url)
	if err != nil {
		url = "http://q4.qlogo.cn/g?b=qq&nk=" + uidStr + "&s=100"
		respBytes, err = net_tools.Download(url)
		if err != nil {
			return nil, err
		}
	}

	redisClient.Set(ctx, uidStr+"avatar", respBytes, 48*time.Hour)
	return respBytes, nil
}

func GetRedisCli() *redis.Client {
	return redisClient
}

// 输入为text里的字体路径
func GetFont(path string) ([]byte, error) {
	a := strings.LastIndex(path, "/")
	if a < 0 {
		panic("can not get font name: " + path)
	}
	name := path[a+1:]

	ctx := context.TODO()
	res, err := redisClient.Get(ctx, name+"font").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if res != "" {
		return []byte(res), nil
	}
	data, err := os.ReadFile(path)
	if err != nil && err != io.EOF {
		return nil, err
	}
	redisClient.Set(ctx, name+"font", data, 7*24*time.Hour)
	return data, nil
}

func GetDefaultFont() ([]byte, error) {
	return GetFont(defalutFontPath)
}
