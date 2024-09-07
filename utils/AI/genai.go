package ai

import (
	"MiniBot/utils/path"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

var (
	config    Config
	dataPath  = path.GetDataPath()
	IM        = IntroduceManger{}
	utilsName = "AI"
)

type session struct {
	timestamp   time.Time
	chatSession *genai.ChatSession
}

type aiBot struct {
	model      *genai.GenerativeModel
	sessionMap map[int64]*session
	sync.RWMutex
}

var AIBot *aiBot

func init() {
	configBytes, err := os.ReadFile(filepath.Join(dataPath, "ai.yaml"))
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}

	IMBytes, err := os.ReadFile(filepath.Join(dataPath, "introduce.json"))
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}
	err = json.Unmarshal(IMBytes, &IM)
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}

	ctx := context.Background()

	key := config.Key
	// proxyURL := "http://192.168.241.1:10811"
	// transport := &http.Transport{
	// 	Proxy: http.ProxyURL(proxyURL),
	// }
	// httpClient := &http.Client{
	// 	Transport: transport,
	// }
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}

	model := client.GenerativeModel("gemini-1.5-flash")

	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockOnlyHigh,
		},
	}

	bot := &aiBot{
		model:      model,
		sessionMap: map[int64]*session{},
	}

	AIBot = bot

	go func() {
		// 每一小时清理会话时长超过两小时的会话
		for range time.NewTicker(1 * time.Hour).C {
			bot.CleanSession()
		}
	}()

}

func (a *aiBot) SendMsg(msg string) (string, error) {
	ctx := context.Background()
	resp, err := a.model.GenerateContent(ctx, genai.Text(msg))
	if err != nil {
		log.Error().Str("name", utilsName).Err(err).Msg("")
		return "", err
	}
	return getResponseString(resp), nil
}

// func printResponse(resp *genai.GenerateContentResponse) {
// 	for _, cand := range resp.Candidates {
// 		if cand.Content != nil {
// 			for _, part := range cand.Content.Parts {
// 				fmt.Println(part)
// 			}
// 		}
// 	}
// 	fmt.Println("---")
// }

func getResponseString(resp *genai.GenerateContentResponse) string {
	res := ""
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				res += fmt.Sprint(part)
			}
		}
	}
	return res
}

func (a *aiBot) SendMsgWithSession(key int64, msg string) (string, error) {
	ctx := context.Background()

	a.RLock()
	session := a.sessionMap[key]
	a.RUnlock()

	// 不存在会话
	if session == nil {
		return "", fmt.Errorf("not session")
	}

	resp, err := session.chatSession.SendMessage(ctx, genai.Text(msg))
	if err != nil {
		log.Error().Str("name", utilsName).Err(err).Msg("")
		if err.Error() == "blocked: candidate: FinishReasonSafety" {
			return "不可以说这种事情，达咩！", nil
		}
		return "", err
	}
	return getResponseString(resp), nil
}

func (a *aiBot) SendPartsWithSession(key int64, parts ...genai.Part) (string, error) {
	ctx := context.Background()

	a.RLock()
	session := a.sessionMap[key]
	a.RUnlock()

	// 不存在会话
	if session == nil {
		return "", fmt.Errorf("not session")
	}

	resp, err := session.chatSession.SendMessage(ctx, parts...)
	if err != nil {
		log.Error().Str("name", utilsName).Err(err).Msg("")
		if err.Error() == "blocked: candidate: FinishReasonSafety" {
			return "不可以说这种事情，达咩！", nil
		}
		return "", err
	}
	return getResponseString(resp), nil
}

func (a *aiBot) CreateSession(key int64, systemInstruction string) *session {
	a.Lock()
	defer a.Unlock()

	a.model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}

	newSession := &session{
		chatSession: a.model.StartChat(),
		timestamp:   time.Now(),
	}

	a.sessionMap[key] = newSession
	return newSession
}

func (a *aiBot) CleanSession() {
	a.Lock()
	defer a.Unlock()
	for k, v := range a.sessionMap {
		if time.Since(v.timestamp) > 2*time.Hour {
			delete(a.sessionMap, k)
		}
	}
}

func (a *aiBot) DelSession(key int64) {
	a.Lock()
	defer a.Unlock()
	delete(a.sessionMap, key)
}
