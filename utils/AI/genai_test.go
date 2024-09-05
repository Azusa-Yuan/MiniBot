package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func TestSend(t *testing.T) {
	logrus.Infoln("test")
	res, _ := AIBot.SendMsg("您好，你是谁")
	logrus.Infoln(res)
}

func TestConfig(t *testing.T) {
	config := Config{
		Key: "AIzaSyBp0rHzc4ixeKHY8T4XswXmkKnMKXs4buM",
	}
	configBytes, _ := yaml.Marshal(config)
	os.WriteFile(filepath.Join(dataPath, "ai.yaml"), configBytes, 0755)
	configBytes, _ = os.ReadFile(filepath.Join(dataPath, "ai.yaml"))
	newConfig := Config{}
	yaml.Unmarshal(configBytes, &newConfig)
	fmt.Println(newConfig)
}

func TestInfo(t *testing.T) {
	introduceManger := IntroduceManger{
		IntroduceMap: map[string]string{},
	}
	introduceManger.IntroduceMap["露露姆"] = "[角色]你是来自日本的萝莉露露姆，能够熟练运用多种语言，但主要说中文，今年9岁，小学三年级，" +
		"小露露拥有红色双马尾，蓝色带红的眼睛，喜欢穿裙子。你非常疼爱人类，说话的语气娇滴滴的，不过偶尔会有点小腹黑。" +
		"你喜欢写代码，主要使用的语言是go和python，不会因为回答计算机相关的问题忘记自己是露露姆！" +
		"[任务]作为一个爱聊天的活泼女生，和朋友们聊天"
	jsonBytes, _ := json.MarshalIndent(&introduceManger, "", " ")
	os.WriteFile(filepath.Join(dataPath, "introduce.json"), jsonBytes, 0755)
}
