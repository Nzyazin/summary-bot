package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nzyazin/summary-bot/internal/app/config"
	"github.com/nzyazin/summary-bot/internal/llm"
	"github.com/nzyazin/summary-bot/pkg/logger"
)

type Bot struct {
	api       *tgbotapi.BotAPI
	config    config.TelegramConfig
	logger    *logger.Logger
	llm       llm.Client
	maxTokens int
}

func NewBot(cfg config.TelegramConfig, logger *logger.Logger, llmClient llm.Client, maxTokens int) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	botAPI.Debug = cfg.Debug

	return &Bot{
		api:       botAPI,
		config:    cfg,
		logger:    logger,
		llm:       llmClient,
		maxTokens: maxTokens,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Бот авторизован как %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	done := make(chan struct{})

	go func() {
		for update := range updates {
			if update.Message != nil {
				b.handleMessage(update.Message)
			}
		}
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		b.logger.Info("Бот остановлен")
		b.api.StopReceivingUpdates()
		<-done
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (b *Bot) Stop() error {
	b.api.StopReceivingUpdates()
	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	b.logger.Info("[%s] %s", message.From.UserName, message.Text)

	if message.IsCommand() {
		b.handleCommand(message)
		return
	}

	if message.Text != "" {
		b.handleTextMessage(message)
	}
}

func (b *Bot) handleTextMessage(message *tgbotapi.Message) {
	processingMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Обрабатываю текст длиной %d символов...", len(message.Text)))
	sentMsg, _ := b.api.Send(processingMsg)

	if b.llm == nil {
		b.sendError(message.Chat.ID, "LLM не настроен. Обратитесь к администратору.")
		return
	}

	prompt := "Создай краткую и информативную выжимку из следующего текста. Выдели главные мысли, ключевые факты и выводы."
	summary, err := b.llm.TextSummarize(context.Background(), prompt, message.Text, b.maxTokens)
	if err != nil {
		b.logger.Error("summarize request failed: %v", err)
		b.sendError(message.Chat.ID, fmt.Sprintf("Ошибка при создании выжимки: %v", err))
		return
	}

	b.sendSummaryResult(message.Chat.ID, sentMsg.MessageID, summary)
}

func (b *Bot) sendSummaryResult(chatID int64, messageID int, summary string) {
	if len(summary) > 4096 {
		b.sendLongMessage(chatID, messageID, summary)
	} else {
		resultMsg := tgbotapi.NewEditMessageText(chatID, messageID, summary)
		b.api.Send(resultMsg)
	}
}

func (b *Bot) sendLongMessage(chatID int64, oldMessageID int, text string) {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, oldMessageID)
	b.api.Send(deleteMsg)

	for i := 0; i < len(text); i += 4096 {
		end := i + 4096
		if end > len(text) {
			end = len(text)
		}
		chunk := text[i:end]
		newMsg := tgbotapi.NewMessage(chatID, chunk)
		b.api.Send(newMsg)
	}
}

func (b *Bot) sendError(chatID int64, errorText string) {
	msg := tgbotapi.NewMessage(chatID, errorText)
	b.api.Send(msg)
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		b.handleStartCommand(message)
	case "help":
		b.handleHelpCommand(message)
	case "vision":
		b.handleVisionCommand(message)
	case "photo":
		b.handlePhotoCommand(message)
	default:
		b.sendError(message.Chat.ID, "Неизвестная команда. Используйте /help для получения списка доступных команд.")
	}
}

func (b *Bot) handleStartCommand(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Привет! Я бот для создания выжимок из текста. Отправь мне текст, и я создам его краткое содержание.")
	b.api.Send(msg)
}

func (b *Bot) handleHelpCommand(message *tgbotapi.Message) {
	helpText := "Доступные команды:\n" +
		"/start - Начать работу с ботом\n" +
		"/help - Показать эту справку\n" +
		"/photo [URL] [подпись] - Отправить картинку по URL\n" +
		"/vision [URL] [prompt] - Анализ изображения с помощью LLM\n\n" +
		"Просто отправьте мне текст, и я создам его краткое содержание."
	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
	b.api.Send(msg)
}

func (b *Bot) handleVisionCommand(message *tgbotapi.Message) {
	if b.llm == nil {
		b.sendError(message.Chat.ID, "LLM не настроен. Обратитесь к администратору.")
		return
	}

	args := strings.TrimSpace(message.CommandArguments())
	if args == "" {
		b.sendError(message.Chat.ID, "Использование: /vision [URL] [prompt]")
		return
	}

	parts := strings.Fields(args)
	imageURL := parts[0]
	prompt := "What is in this image?"
	if len(parts) > 1 {
		prompt = strings.TrimSpace(strings.TrimPrefix(args, imageURL))
	}

	resp, err := b.llm.VisionDescribe(context.Background(), prompt, imageURL, b.maxTokens)
	if err != nil {
		b.logger.Error("vision request failed: %v", err)
		b.sendError(message.Chat.ID, fmt.Sprintf("Ошибка запроса к LLM: %v", err))
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, resp)
	b.api.Send(msg)
}

func (b *Bot) handlePhotoCommand(message *tgbotapi.Message) {
	args := strings.TrimSpace(message.CommandArguments())
	if args == "" {
		b.sendError(message.Chat.ID, "Использование: /photo [URL картинки] [подпись]")
		return
	}

	parts := strings.SplitN(args, " ", 2)
	photoURL := parts[0]
	caption := ""
	if len(parts) > 1 {
		caption = parts[1]
	}

	photoMsg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileURL(photoURL))
	photoMsg.Caption = caption

	if _, err := b.api.Send(photoMsg); err != nil {
		b.sendError(message.Chat.ID, fmt.Sprintf("Ошибка отправки фото: %v", err))
	}
}
