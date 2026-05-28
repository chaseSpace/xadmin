package alertbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	systemrepo "monorepo/internal/repo/system"
	"monorepo/pkg/logger"

	"go.uber.org/zap"
)

// Bot is the interface for sending alert messages.
type Bot interface {
	Send(ctx context.Context, message string) error
	Type() string
}

// TelegramBot sends messages via Telegram Bot API.
type TelegramBot struct {
	Name      string
	Username  string // chat_id
	Token     string
	ParseMode string // 解析模式：空(默认纯文本) / Markdown / MarkdownV2 / HTML
}

func (b *TelegramBot) Type() string { return "telegram" }

func (b *TelegramBot) Send(ctx context.Context, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.Token)
	// 将模板中的字面 \n 转为实际换行符，确保 TG 正确渲染
	message = strings.ReplaceAll(message, `\n`, "\n")
	params := url.Values{
		"chat_id": {b.Username},
		"text":    {message},
	}
	// parse_mode: 空=不传, tg_default=不传, Markdown, MarkdownV2, HTML
	if b.ParseMode != "" && b.ParseMode != "tg_default" {
		params.Set("parse_mode", b.ParseMode)
	}
	logger.Info("telegram send request",
		zap.String("chat_id", b.Username),
		zap.String("parse_mode", b.ParseMode),
		zap.Int("text_len", len(message)),
	)
	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		logger.Error("telegram send network error", zap.Error(err))
		return fmt.Errorf("telegram send failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Warn("telegram send failed", zap.Int("status", resp.StatusCode), zap.String("chat_id", b.Username))
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("telegram returned 404, token可能无效")
		}
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("telegram returned 401, token无效")
		}
		if resp.StatusCode == http.StatusBadRequest {
			return fmt.Errorf("telegram returned 400, 请确保 chat_id 有效")
		}
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}
	logger.Info("telegram send success", zap.String("chat_id", b.Username))
	return nil
}

// FeishuBot sends messages via Feishu webhook.
type FeishuBot struct {
	Name      string
	Token     string
	ParseMode string // 空=text, post=富文本, interactive=卡片
}

func (b *FeishuBot) Type() string { return "feishu" }

func (b *FeishuBot) Send(ctx context.Context, message string) error {
	apiURL := fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", b.Token)

	var body map[string]any
	switch b.ParseMode {
	case "post":
		var postContent any
		if err := json.Unmarshal([]byte(message), &postContent); err != nil {
			return fmt.Errorf("feishu post content invalid JSON: %w", err)
		}
		body = map[string]any{"msg_type": "post", "content": map[string]any{"post": postContent}}
	case "interactive":
		var card any
		if err := json.Unmarshal([]byte(message), &card); err != nil {
			return fmt.Errorf("feishu interactive card invalid JSON: %w", err)
		}
		body = map[string]any{"msg_type": "interactive", "card": card}
	default:
		message = strings.ReplaceAll(message, `\n`, "\n")
		body = map[string]any{"msg_type": "text", "content": map[string]string{"text": message}}
	}

	data, _ := json.Marshal(body)
	logger.Info("feishu send request", zap.String("parse_mode", b.ParseMode), zap.Int("body_len", len(data)), zap.String("body", string(data)))
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(data))
	if err != nil {
		logger.Error("feishu send network error", zap.Error(err))
		return fmt.Errorf("feishu send failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Warn("feishu send failed", zap.Int("status", resp.StatusCode))
		return fmt.Errorf("feishu returned status %d", resp.StatusCode)
	}
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.Code != 0 {
		logger.Warn("feishu send biz error", zap.Int("code", result.Code), zap.String("msg", result.Msg))
		hint := ""
		if result.Code == 19002 {
			hint = "，请确保场景的解析模式与模板内容匹配"
		}
		return fmt.Errorf("feishu error: %s (code=%d)%s", result.Msg, result.Code, hint)
	}
	logger.Info("feishu send success")
	return nil
}

// GetBots returns callable Bot instances for a given scene key.
// It queries scenes by key, then loads the associated bot.
func GetBots(ctx context.Context, sceneKey string) []Bot {
	repo := systemrepo.NewRepo()
	scenes, err := repo.FindScenesByKey(ctx, sceneKey)
	if err != nil {
		logger.Error("Failed to find scenes", zap.String("scene", sceneKey), zap.Error(err))
		return nil
	}
	bots := make([]Bot, 0, len(scenes))
	for _, scene := range scenes {
		if scene.BotID <= 0 {
			continue
		}
		bot, err := repo.GetAlertBot(ctx, scene.BotID)
		if err != nil || !bot.Enabled {
			continue
		}
		switch bot.BotType {
		case "telegram":
			bots = append(bots, &TelegramBot{Name: bot.Name, Username: scene.GroupID, Token: bot.Token, ParseMode: scene.ParseMode})
		case "feishu":
			bots = append(bots, &FeishuBot{Name: bot.Name, Token: bot.Token, ParseMode: scene.ParseMode})
		}
	}
	return bots
}

// Notify sends a message to all bots matching the scene key.
func Notify(ctx context.Context, sceneKey, message string) {
	for _, bot := range GetBots(ctx, sceneKey) {
		if err := bot.Send(ctx, message); err != nil {
			logger.Error("Alert bot send failed",
				zap.String("bot_type", bot.Type()),
				zap.String("scene", sceneKey),
				zap.Error(err),
			)
		}
	}
}
