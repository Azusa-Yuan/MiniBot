package meme

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestCreatImg(t *testing.T) {
	// data := map[string]interface{}{
	// 	"texts": []string{"hello", "world"},
	// }
	// jsondata, _ := json.Marshal(data)
	postData := url.Values{}
	postData.Add("texts", "hello")
	postData.Add("texts", "world")
	resp, _ := http.DefaultClient.Post(baseUrl+"5000choyen/", "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
	res, _ := io.ReadAll(resp.Body)
	os.WriteFile("./test.jpg", res, 0644)
}

func TestGetList(t *testing.T) {
	list, _ := GetList()
	fmt.Println(list)
	fmt.Println(len(list))
}

func TestGetInfo(t *testing.T) {
	fmt.Println(GetEmojiInfo("fencing"))
}

func TestGetAlwaysInfo(t *testing.T) {
	info, _ := GetEmojiInfo("always")
	fmt.Println(info)
	fmt.Println(info.Shortcuts)
	fmt.Println(info.ParamsType.ArgsType.ArgsExamples...)
	fmt.Println(info.ParamsType.ArgsType.ParserOptions)
	parserOptions := info.ParamsType.ArgsType.ParserOptions
	for _, parserOption := range parserOptions {
		for _, arg := range parserOption.Args {
			fmt.Println("--"+arg.Name, parserOption.HelpText)
		}
	}
}

func TestQueryEmojiInfo(t *testing.T) {
	fmt.Println(QueryEmojiInfo("always"))
}

func TestCreateAlways(t *testing.T) {
	file, _ := os.Open("./lulumu.bmp")
	data, _ := io.ReadAll(file)
	args := map[string]any{}
	args["user_infos"] = []UserInfo{}
	args["mode"] = "loop"
	argsData, _ := json.Marshal(args)
	fmt.Println(string(argsData))
	data, err := CreateEmoji("always", [][]byte{data}, nil, string(argsData))
	os.WriteFile("./always.gif", data, 0644)
	fmt.Println(err)
	args["mode"] = "circle"
	argsData, _ = json.Marshal(args)
	data, err = CreateEmoji("always", [][]byte{data}, nil, string(argsData))
	fmt.Println(err)
	os.WriteFile("./always_circle.gif", data, 0644)
}

func TestCreateLittleAngel(t *testing.T) {
	file, _ := os.Open("./lulumu.bmp")
	data, _ := io.ReadAll(file)
	args := map[string]any{}
	args["user_infos"] = []UserInfo{
		{Name: "露露姆", Gender: "male"},
	}
	argsData, _ := json.Marshal(args)
	fmt.Println(string(argsData))
	data, err := CreateEmoji("little_angel", [][]byte{data}, nil, string(argsData))
	os.WriteFile("./little_angel.jpg", data, 0644)
	fmt.Println(err)
}

func TestArgs(t *testing.T) {

	input := "一直一直   --mode   normal  --mode nooo"
	pattern := `--(\S+)\s+(\S+)`
	reArgs := regexp.MustCompile(pattern)
	matchs := reArgs.FindAllStringSubmatch(input, -1)
	fmt.Println(len(matchs))
	fmt.Println(matchs)
	res := reArgs.ReplaceAllString(input, "")
	fmt.Println(res)
}

func TestCreateEmoji3(t *testing.T) {
	data, err := CreateEmoji("fanatic", [][]byte{}, []string{"露露姆"}, "")
	os.WriteFile("./fanatic.jpg", data, 0644)
	fmt.Println(err)
}

func TestCreateRBQ(t *testing.T) {
	file, _ := os.Open("./lulumu.bmp")
	data, _ := io.ReadAll(file)
	data, err := CreateEmoji("rbq", [][]byte{data}, []string{"露露姆"}, "")
	os.WriteFile("./rbq.jpg", data, 0644)
	fmt.Println(err)
}

func TestCreateBudong(t *testing.T) {
	file, _ := os.Open("./lulumu.bmp")
	data, _ := io.ReadAll(file)
	data, err := CreateEmoji("budong", [][]byte{data}, nil, "")
	os.WriteFile("./budong.jpg", data, 0644)
	fmt.Println(err)
}

func TestInit(t *testing.T) {
	InitMeme()
	fmt.Println(emojiMap)
	fmt.Println(len(emojiMap))
}

func TestHelp(t *testing.T) {
	InitMeme()
	data, err := GetHelp()
	fmt.Println(err)
	os.WriteFile("./help.jpg", data, 0644)
}
