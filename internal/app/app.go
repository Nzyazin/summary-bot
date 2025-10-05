package app

import (
    "context"
    "fmt"

    "github.com/nzyazin/summary-bot/internal/app/config"
    "github.com/nzyazin/summary-bot/internal/llm"
    "github.com/nzyazin/summary-bot/pkg/logger"
    "github.com/nzyazin/summary-bot/internal/bot"
)

type App struct {
    cfg      *config.Config
    bot      *bot.Bot
    logger   *logger.Logger
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
    log, err := logger.NewLogger(cfg.Log.Level, cfg.Log.File)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize logger: %w", err)
    }

    // Initialize LLM client if configured
    var llmClient llm.Client
    switch cfg.LLM.Provider {
    case "openrouter":
        oc, err := llm.NewOpenRouterClient(cfg.LLM.APIKey, cfg.LLM.BaseURL, cfg.LLM.ModelName)
        if err != nil {
            return nil, fmt.Errorf("failed to initialize openrouter client: %w", err)
        }
        llmClient = oc
    case "artemox":
        ac, err := llm.NewArtemoxClient(cfg.LLM.APIKey, cfg.LLM.ModelName)
        if err != nil {
            return nil, fmt.Errorf("failed to initialize artemox client: %w", err)
        }
        llmClient = ac
    }

    maxTokens := cfg.LLM.MaxTokens
    if maxTokens <= 0 {
        maxTokens = 500
    }

    telegramBot, err := bot.NewBot(cfg.Telegram, log, llmClient, maxTokens)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize telegram bot: %w", err)
    }

    return &App{
        cfg:      cfg,
        bot:      telegramBot,
        logger:   log,
    }, nil
}

func (a *App) Run(ctx context.Context) error {
	return a.bot.Start(ctx)
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.bot.Stop(); err != nil {
		return fmt.Errorf("error stopping bot: %w", err)
	}

	return nil
}
