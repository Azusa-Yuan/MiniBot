package cache

import (
	"MiniBot/utils"
	"MiniBot/utils/net_tools"
	"MiniBot/utils/path"
	"MiniBot/utils/text"
	zero "ZeroBot"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/redis/go-redis/v9"
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
	utilsName       = "redis"
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
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}

	redisClientTmp := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Rdb.Host, config.Rdb.Port),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	// Ping Redis server
	_, err = redisClientTmp.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}
	log.Info().Str("name", utilsName).Msg("connect redis success!")
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
	url := "https://q4.qlogo.cn/g?b=qq&nk=" + uidStr + "&s=640"
	respBytes, err := net_tools.Download(url)
	if err != nil {
		url = "https://q4.qlogo.cn/g?b=qq&nk=" + uidStr + "&s=100"
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
		return utils.StringToBytes(res), nil
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

func GetGroupMemberList(bid, gid int64) ([]int64, error) {
	key := strconv.FormatInt(gid, 10) + "mem_list"
	res, err := redisClient.Get(context.TODO(), key).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	memberList := []int64{}
	if res != "" {
		json.Unmarshal(utils.StringToBytes(res), &memberList)
		return memberList, nil
	}

	bot, err := zero.GetBot(bid)
	if err != nil {
		return nil, err
	}

	tempRaw, err := bot.GetGroupMemberListNoCache(gid)
	if err != nil {
		return nil, err
	}
	temp := tempRaw.Array()
	sort.SliceStable(temp, func(i, j int) bool {
		return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
	})
	for _, v := range temp {
		memberList = append(memberList, v.Get("user_id").Int())
	}

	data, err := json.Marshal(memberList)
	if err != nil {
		return nil, err
	}

	redisClient.Set(context.TODO(), key, data, time.Hour*24)
	return memberList, nil
}
