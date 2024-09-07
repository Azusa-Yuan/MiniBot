package pcrjjc3

import (
	"MiniBot/utils/path"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog/log"
)

// Rank struct represents a row in the rank CSV file
type Rank struct {
	Rank string
	Exp  string
}

// var pcrClient *pcrclient
var (
	dataPath = path.GetPluginDataPath()
	// proxy   = "http://192.168.241.1:10811"
	proxy = ""
	cxMap = map[string]string{
		"1": "美食殿堂",
		"2": "真步真步王国",
		"3": "破晓之星",
		"4": "小小甜心",
	}
	clientMap = map[string]*pcrclient{}
	header    = map[string]string{}
	ranks     []Rank
)

func getClient(cx string) *pcrclient {
	// 拼接文件路径
	cxPath := filepath.Join(dataPath, fmt.Sprintf("%scx_tw.sonet.princessconnect.v2.playerprefs.xml", cx))

	// 检查文件是否存在
	if _, err := os.Stat(cxPath); os.IsNotExist(err) {
		return nil
	}

	infoMap := decryptxml(cxPath)
	client, err := CreatePcrclient(infoMap["UDID"], infoMap["SHORT_UDID_lowBits"],
		infoMap["VIEWER_ID_lowBits"], infoMap["TW_SERVER_ID"], proxy, header)
	if err != nil {
		return nil
	}

	return client
}

func init() {
	// 获取header
	headerReader, _ := os.Open(filepath.Join(dataPath, "headers.json"))
	raw, _ := io.ReadAll(headerReader)

	err := json.Unmarshal(raw, &header)
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
		return
	}

	// 创建客户端
	client := getClient("1")
	if client != nil {
		clientMap["1"] = client
	}
	client = getClient("2")
	if client != nil {
		clientMap["2"] = client
		clientMap["3"] = client
		clientMap["4"] = client
	}

	// 获取用户信息
	userFile, _ := os.Open(filepath.Join(dataPath, "binds.json"))
	userdata, _ := io.ReadAll(userFile)

	json.Unmarshal(userdata, &userInfoManage)

	// 导入exp经验等级表
	rankExpPath := dataPath + "/rank_exp.csv" // Update this path to the actual CSV file path
	ranks, err = loadRankExp(rankExpPath)
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
		return
	}
}

// loadRankExp reads the CSV file and returns a slice of Rank structs
func loadRankExp(path string) ([]Rank, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Read all records from the CSV
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file does not contain enough data")
	}

	// Extract headers
	// headers := records[0]
	ranks := make([]Rank, 0, len(records)-1)

	// Read rows and convert to Rank structs
	for _, record := range records[1:] {
		if len(record) < 2 {
			continue
		}
		rank := record[0]
		exp := record[1]
		ranks = append(ranks, Rank{Rank: rank, Exp: exp})
	}

	return ranks, nil
}

// calKnightRank calculates the rank based on the target value
func calKnightRank(targetValue int) int {
	var targetRank int = 1

	for _, row := range ranks {
		exp, err := strconv.Atoi(row.Exp)
		if err != nil {
			log.Error().Str("name", pluginName).Err(err).Msg("")
			return targetRank
		}

		if targetValue >= exp {
			rank, err := strconv.Atoi(row.Rank)
			if err != nil {
				log.Error().Str("name", pluginName).Err(err).Msg("")
				return targetRank
			}
			targetRank = rank
		} else {
			break
		}
	}

	return targetRank
}

func calculateDomain(sum int) string {
	// 大关卡
	big := (sum-1)/10 + 1
	// 小关卡
	small := sum % 10
	if small == 0 && big > 0 {
		small = 10
	}
	return fmt.Sprintf("%d-%d", big, small)
}

func savaHeader() error {
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dataPath, "headers.json"), headerJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}
