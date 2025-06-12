package presentation

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/kakudo415/survey-bot/usecase"
)

type SlackEventController struct {
	eventHandler  *usecase.SlackEventHandler
	signingSecret string
}

func NewSlackEventController(eventHandler *usecase.SlackEventHandler, signingSecret string) *SlackEventController {
	return &SlackEventController{
		eventHandler:  eventHandler,
		signingSecret: signingSecret,
	}
}

func (c *SlackEventController) HandleSlackEvent(w http.ResponseWriter, r *http.Request) {
	// リクエストボディを読み取り
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Slack署名検証
	if !c.verifySlackSignature(r, body) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// JSONをパース
	var event usecase.SlackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// URL verification challenge
	if event.Type == "url_verification" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(event.Challenge))
		return
	}

	// イベント処理
	if err := c.eventHandler.HandleEvent(r.Context(), event); err != nil {
		// ログに記録するが、Slackには200を返す（再送を避けるため）
		fmt.Printf("Failed to handle event: %v\n", err)
	}

	w.WriteHeader(http.StatusOK)
}

func (c *SlackEventController) verifySlackSignature(r *http.Request, body []byte) bool {
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	signature := r.Header.Get("X-Slack-Signature")

	if timestamp == "" || signature == "" {
		return false
	}

	// タイムスタンプが古すぎる場合は拒否（リプレイ攻撃対策）
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	
	if time.Now().Unix()-ts > 300 { // 5分以内
		return false
	}

	// 署名を計算
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(c.signingSecret))
	mac.Write([]byte(baseString))
	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	// 署名を比較
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}