package meme

import (
	"MiniBot/utils/cache"
	"bytes"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"

	"github.com/FloatTech/gg"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var (
	client       = &http.Client{}
	baseUrl      = "http://127.0.0.1:2233/memes/"
	emojiMap     = map[string]string{}
	emojiInfoMap = map[string]*EmojiInfo{}
	cmdList      = []string{}
	colnum       = 3
	helpData     []byte
)

func GetHelp() ([]byte, error) {

	if len(helpData) > 0 {
		return helpData, nil
	}

	number := len(cmdList) / colnum
	if len(cmdList)%colnum > 0 {
		number++
	}
	fontSize := 30.0
	canvas := gg.NewContext(1500, int(220+fontSize*float64(number)))
	canvas.SetRGB(1, 1, 1) // 白色
	canvas.Clear()
	/***********获取字体，可以注销掉***********/
	data, err := cache.GetDefaultFont()
	if err != nil {
		return nil, err
	}
	/***********设置字体颜色为黑色***********/
	canvas.SetRGB(0, 0, 0)
	/***********设置字体大小,并获取字体高度用来定位***********/
	if err = canvas.ParseFontFace(data, fontSize*2); err != nil {
		return nil, err
	}
	sl, h := canvas.MeasureString("表情包列表")
	/***********绘制标题***********/
	canvas.DrawString("表情包列表", (1500-sl)/2, 140-1.2*h) // 放置在中间位置
	/***********设置字体大小,并获取字体高度用来定位***********/
	if err = canvas.ParseFontFace(data, 1.5*fontSize); err != nil {
		return nil, err
	}

	_, h = canvas.MeasureString("焯")
	// 打印数据
	for i := 0; i < len(cmdList); i += colnum {

		for j := i; j < min(len(cmdList), i+colnum); j++ {
			canvas.DrawString(cmdList[j], float64(j%colnum*1500/colnum), 180+fontSize*float64(i)-h)
		}

	}
	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024*4))
	err = jpeg.Encode(buffer, canvas.Image(), &jpeg.Options{Quality: 70})
	if err != nil {
		return nil, err
	}

	data = buffer.Bytes()
	helpData = data
	return data, nil
}

func GetList() ([]string, error) {
	resp, err := client.Get(baseUrl + "keys")
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
	resp, err := client.Get(baseUrl + key + "/info")
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
	url := baseUrl + emojiPath + "/"
	logrus.Info("creat emoji:", url)
	req, err := http.NewRequest("POST", url, &requestBody)
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
		logrus.Error(err)
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
					logrus.Error(err)
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

func fastJudge(path string, imgLen int, textLen int) bool {
	if info, ok := emojiInfoMap[path]; ok {
		if !(imgLen >= int(info.ParamsType.MinImages) && imgLen <= int(info.ParamsType.MaxImages)) {
			logrus.Debugln("图片数量不对", imgLen, info.ParamsType.MinImages, info.ParamsType.MaxImages)
			return false
		}
		if textLen == 0 {
			return true
		}
		if textLen >= int(info.ParamsType.MinTexts) && textLen <= int(info.ParamsType.MaxTexts) {
			return true
		}
		logrus.Debugln("文字长度不对", textLen, info.ParamsType.MinTexts, info.ParamsType.MaxTexts)
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
	resText := key + "帮助如下:"
	resText += fmt.Sprintf("\n 最小文本信息个数:%d, 最大文本信息个数:%d", info.ParamsType.MinTexts, info.ParamsType.MaxTexts)
	resText += fmt.Sprintf("\n 最小图片张数:%d, 最大图片张数:%d", info.ParamsType.MinImages, info.ParamsType.MaxImages)
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
