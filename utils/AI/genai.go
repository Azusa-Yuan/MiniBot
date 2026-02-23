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

	"github.com/rs/zerolog/log"
	"google.golang.org/genai"
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
	chatSession *genai.Chat
}

type aiBot struct {
	client     *genai.Client
	modelName  string
	config     *genai.GenerateContentConfig
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

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  config.Key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}

	modelName := "gemini-2.5-flash"
	cfg := &genai.GenerateContentConfig{
		SafetySettings: []*genai.SafetySetting{
			{
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockThresholdBlockOnlyHigh,
			},
			{
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockThresholdBlockOnlyHigh,
			},
		},
	}

	bot := &aiBot{
		client:     client,
		modelName:  modelName,
		config:     cfg,
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
	contents := []*genai.Content{
		{Parts: []*genai.Part{{Text: msg}}},
	}
	resp, err := a.client.Models.GenerateContent(ctx, a.modelName, contents, a.config)
	if err != nil {
		log.Error().Str("name", utilsName).Err(err).Msg("")
		return "", err
	}
	return getResponseString(resp), nil
}

func getResponseString(resp *genai.GenerateContentResponse) string {
	res := ""
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if part != nil {
					res += part.Text
				}
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

	if session == nil {
		return "", fmt.Errorf("not session")
	}

	resp, err := session.chatSession.SendMessage(ctx, genai.Part{Text: msg})
	if err != nil {
		log.Error().Str("name", utilsName).Err(err).Msg("")
		if err.Error() == "blocked: candidate: FinishReasonSafety" || err.Error() == "blocked: prompt: BlockReasonOther" {
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

	cfg := &genai.GenerateContentConfig{
		SafetySettings: a.config.SafetySettings,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemInstruction}},
		},
	}

	chatSession, err := a.client.Chats.Create(context.Background(), a.modelName, cfg, nil)
	if err != nil {
		log.Error().Str("name", utilsName).Err(err).Msg("")
		return nil
	}

	newSession := &session{
		chatSession: chatSession,
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
