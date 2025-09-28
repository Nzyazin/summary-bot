package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nzyazin/summary-bot/internal/app/config"
	"github.com/nzyazin/summary-bot/pkg/logger"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	config config.TelegramConfig
	logger *logger.Logger
}

func NewBot(cfg config.TelegramConfig, logger *logger.Logger) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	botAPI.Debug = cfg.Debug

	return &Bot{
		api:    botAPI,
		config: cfg,
		logger: logger,
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
		summary := fmt.Sprintf("Получен текст длиной %d символов.\n\nВ будущем здесь будет выжимка, созданная с помощью LLM.", len(message.Text))
		msg := tgbotapi.NewMessage(message.Chat.ID, summary)
		b.api.Send(msg)
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
			"Просто отправьте мне текст, и я создам его краткое содержание."
	default:
		msg.Text = "Неизвестная команда. Используйте /help для получения списка доступных команд."
	}

	b.api.Send(msg)
}
