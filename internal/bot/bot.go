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
		// Отправляем сообщение о начале обработки
		processingMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Обрабатываю текст длиной %d символов...", len(message.Text)))
		sentMsg, _ := b.api.Send(processingMsg)

		// Проверяем, настроен ли LLM клиент
		if b.llm == nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "LLM не настроен. Обратитесь к администратору.")
			_, _ = b.api.Send(msg)
			return
		}

		// Создаем промпт для выжимки
		prompt := "Создай краткую и информативную выжимку из следующего текста. Выдели главные мысли, ключевые факты и выводы."

		// Отправляем запрос к LLM
		summary, err := b.llm.TextSummarize(context.Background(), prompt, message.Text, b.maxTokens)
		if err != nil {
			b.logger.Error("summarize request failed: %v", err)
			errorMsg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Ошибка при создании выжимки: %v", err))
			b.api.Send(errorMsg)
			return
		}

		// Отправляем результат
		resultMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sentMsg.MessageID, summary)
		b.api.Send(resultMsg)
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	switch message.Command() {
	case "start":
		msg.Text = "Привет! Я бот для создания выжимок из текста. Отправь мне текст, и я создам его краткое содержание."
	case "help":
		msg.Text = "Доступные команды:\n" +
			"/start - Начать работу с ботом\n" +
			"/help - Показать эту справку\n\n" +
			"Просто отправьте мне текст, и я создам его краткое содержание.\n" +
			"/vision - Выжимка поста с помощью LLM"
	case "vision":
		if b.llm == nil {
			msg.Text = "LLM не настроен. Обратитесь к администратору."
			b.api.Send(msg)
			return
		}
		args := strings.TrimSpace(message.CommandArguments())
		if args == "" {
			msg.Text = "Использование: /vision [prompt]"
			b.api.Send(msg)
			return
		}
		parts := strings.Fields(args)
		imageURL := parts[0]
		var prompt string
		if len(parts) > 1 {
			prompt = strings.TrimSpace(strings.TrimPrefix(args, imageURL))
		}
		if prompt == "" {
			prompt = "What is in this image?"
		}
		resp, err := b.llm.VisionDescribe(context.Background(), prompt, imageURL, b.maxTokens)
		if err != nil {
			b.logger.Error("vision request failed: %v", err)
			msg.Text = fmt.Sprintf("Ошибка запроса к LLM: %v", err)
		} else {
			msg.Text = resp
		}
	default:
		msg.Text = "Неизвестная команда. Используйте /help для получения списка доступных команд."
	}

	b.api.Send(msg)
}
