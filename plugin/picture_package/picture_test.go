package picturepackage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

func TestPage(t *testing.T) {
	maxCgPageNumber, _, _ := initPageNumber()
	for i := 1; i <= maxCgPageNumber; i++ {
		err := getPicID(i, cgType)
		if err != nil {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Println(cgIDList)
	fmt.Println(len(cgIDList))
}

func TestLolicon(t *testing.T) {

	jsonData, err := json.Marshal(map[string]any{
		"size": "regular",
		"tag":  "刻晴",
	})
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Post("https://api.lolicon.app/setu/v2", "application/json", bytes.NewBuffer(jsonData))
	log.Error().Err(err).Msg("")
	if err == nil && resp.StatusCode == http.StatusOK {
		respData, err := io.ReadAll(resp.Body)
		log.Error().Err(err).Msg("")
		if err == nil {
			fmt.Println(gjson.ParseBytes(respData).Get("data"))
			dataArray := gjson.ParseBytes(respData).Get("data").Array()
			if len(dataArray) != 0 {
				imgData := dataArray[0]
				url := imgData.Get("urls").Get("original").String()
				fmt.Println(url)
			}
		}
	}
}
