package meme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

var (
	client       = &http.Client{}
	baseUrl      = "http://127.0.0.1:2233/memes"
	emojiMap     = map[string]string{}
	emojiInfoMap = map[string]*EmojiInfo{}
	cmdList      = []string{}
	helpData     []byte
)

func GetHelp() ([]byte, error) {
	if len(helpData) > 0 {
		return helpData, nil
	}
	u, _ := url.JoinPath(baseUrl, "render_list")
	resp, err := client.Post(u, "application/json", nil)
	if err != nil {
		return nil, err
	}

	// 接收数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("create image: %s, %s", resp.Status, string(data))
	}

	helpData = data
	return data, nil
}

func GetList() ([]string, error) {
	u, _ := url.JoinPath(baseUrl, "keys")
	resp, err := client.Get(u)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	emojiList := []string{}
	err = json.Unmarshal(data, &emojiList)
	if err != nil {
		return nil, err
	}
	return emojiList, nil
}

func GetEmojiInfo(key string) (EmojiInfo, error) {
	emojiInfo := EmojiInfo{}
	u, _ := url.JoinPath(baseUrl, key, "info")
	resp, err := client.Get(u)
	if err != nil {
		return emojiInfo, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return emojiInfo, err
	}
	gjson.ParseBytes(data)
	err = json.Unmarshal(data, &emojiInfo)
	if err != nil {
		return emojiInfo, err
	}
	// fmt.Println(string(data))
	return emojiInfo, nil
}

func CreateEmoji(emojiPath string, images [][]byte, texts []string, args string) ([]byte, error) {
	// 创建一个新的缓冲区
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	var err error

	// 添加images
	for i, image := range images {
		part, err := writer.CreateFormFile("images", fmt.Sprintf("image%d", i))
		if err != nil {
			return nil, err
		}
		_, err = part.Write(image)
		if err != nil {
			return nil, err
		}
	}

	// 添加texts
	for _, text := range texts {
		err = writer.WriteField("texts", text)
		if err != nil {
			return nil, err
		}
	}

	// 添加args
	if args != "" {
		err = writer.WriteField("args", args)
		if err != nil {
			return nil, err
		}
	}

	// 关闭写入器，写入结束边界
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// 创建请求
	u, _ := url.JoinPath(baseUrl, emojiPath)
	log.Info().Str("name", pluginName).Msgf("creat emoji:%s", u)
	req, err := http.NewRequest("POST", u, &requestBody)
	if err != nil {
		return nil, err
	}

	if len(images) > 0 || len(texts) > 0 {
		req.Header.Set("Content-Type", writer.FormDataContentType())
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 接收数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("create image: %s, %s", resp.Status, string(data))
	}
	// fmt.Println(string(data))
	// os.WriteFile("./test.gif", data, 0644)

	return data, nil
}

func InitMeme() error {
	pathList, err := GetList()
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
		return err
	}

	// 多协程加载emoji数据
	wg := sync.WaitGroup{}
	wg.Add(5)
	lock := &sync.RWMutex{}
	ch := make(chan string, 200)
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			path := ""
			for path = range ch {
				emojiInfo, err := GetEmojiInfo(path)
				if err != nil {
					log.Error().Str("name", pluginName).Err(err).Msg("")
					return
				}

				lock.Lock()
				emojiInfoMap[path] = &emojiInfo
				// 整合key和shortcut
				shortCutKey := []string{}
				for _, shortcut := range emojiInfo.Shortcuts {
					shortCutKey = append(shortCutKey, shortcut.Key)
				}
				emojiInfo.Keywords = append(emojiInfo.Keywords, shortCutKey...)

				for _, key := range emojiInfo.Keywords {
					emojiMap[key] = path
				}

				cmdList = append(cmdList, strings.Join(emojiInfo.Keywords[:min(4, len(emojiInfo.Keywords))], "/"))
				lock.Unlock()
			}
		}()
	}

	for _, path := range pathList {
		ch <- path
	}

	close(ch)
	wg.Wait()
	return nil
}

func truncateList(path string, imgStrs []string, texts []string) ([]string, []string) {
	imgLen := len(imgStrs)
	textLen := len(texts)
	if info, ok := emojiInfoMap[path]; ok {
		if imgLen > int(info.ParamsType.MaxImages) {
			imgStrs = imgStrs[:int(info.ParamsType.MaxImages)]
		}
		if textLen > int(info.ParamsType.MaxTexts) {
			texts = texts[:int(info.ParamsType.MaxTexts)]
		}
	}
	return imgStrs, texts
}

func fastJudge(path string, imgStrs []string, texts []string) bool {
	imgLen := len(imgStrs)
	textLen := len(texts)
	if info, ok := emojiInfoMap[path]; ok {
		if !(imgLen >= int(info.ParamsType.MinImages) && imgLen <= int(info.ParamsType.MaxImages)) {
			log.Debug().Str("name", pluginName).Msgf("图片数量不对 actual:%d max:%d min:%d", imgLen, info.ParamsType.MinImages, info.ParamsType.MaxImages)
			return false
		}
		if textLen == 0 {
			return true
		}
		if textLen >= int(info.ParamsType.MinTexts) && textLen <= int(info.ParamsType.MaxTexts) {
			return true
		}
		log.Debug().Str("name", pluginName).Msgf("文字长度不对 actual:%d max:%d min:%d", imgLen, info.ParamsType.MinImages, info.ParamsType.MaxImages)
	}
	return false
}

func QueryEmojiInfo(key string) string {
	var info *EmojiInfo
	info = emojiInfoMap[key]
	if info == nil {
		if path, ok := emojiMap[key]; ok {
			info = emojiInfoMap[path]
		}
	}
	if info == nil {
		return "没有该表情的信息"
	}
	resText := ""
	resText += fmt.Sprintf("\n表情名:%s", info.Key)
	resText += fmt.Sprintf("\n关键词:%s", strings.Join(info.Keywords, "/"))
	if info.ParamsType.MinTexts != info.ParamsType.MaxTexts {
		resText += fmt.Sprintf("\n需要文本数目:%d-%d", info.ParamsType.MinTexts, info.ParamsType.MaxTexts)
	} else {
		resText += fmt.Sprintf("\n需要文本数目:%d", info.ParamsType.MinTexts)
	}
	if info.ParamsType.MinImages != info.ParamsType.MaxImages {
		resText += fmt.Sprintf("\n需要图片数目:%d-%d", info.ParamsType.MinImages, info.ParamsType.MaxImages)
	} else {
		resText += fmt.Sprintf("\n需要图片数目:%d", info.ParamsType.MinImages)
	}
	resText += "\n可选参数如下:"
	parserOptions := info.ParamsType.ArgsType.ParserOptions
	for _, parserOption := range parserOptions {
		for _, arg := range parserOption.Args {
			resText += fmt.Sprint("\n--"+arg.Name, ": ", parserOption.HelpText)
			break
		}
	}
	return resText
}
