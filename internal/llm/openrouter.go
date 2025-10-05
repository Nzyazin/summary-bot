package llm

import (
    "context"
    "errors"
    "fmt"

    openai "github.com/sashabaranov/go-openai"
)

type OpenRouterClient struct {
    client   *openai.Client
    model    string
}

func NewOpenRouterClient(apiKey, baseURL, model string) (*OpenRouterClient, error) {
    if apiKey == "" {
        return nil, errors.New("openrouter api key is empty")
    }
    if baseURL == "" {
        baseURL = "https://openrouter.ai/api/v1"
    }

    cfg := openai.DefaultConfig(apiKey)
    cfg.BaseURL = baseURL

    c := openai.NewClientWithConfig(cfg)
    return &OpenRouterClient{client: c, model: model}, nil
}

func (c *OpenRouterClient) VisionDescribe(ctx context.Context, prompt, imageURL string, maxTokens int) (string, error) {
    if imageURL == "" {
        return "", errors.New("image url is required")
    }

    parts := []openai.ChatMessagePart{
        {Type: openai.ChatMessagePartTypeText, Text: prompt},
        {
            Type: openai.ChatMessagePartTypeImageURL,
            ImageURL: &openai.ChatMessageImageURL{URL: imageURL},
        },
    }

    req := openai.ChatCompletionRequest{
        Model: c.model,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:         openai.ChatMessageRoleUser,
                MultiContent: parts,
            },
        },
    }
    if maxTokens > 0 {
        req.MaxTokens = maxTokens
    }

    resp, err := c.client.CreateChatCompletion(ctx, req)
    if err != nil {
        return "", fmt.Errorf("openrouter vision request failed: %w", err)
    }

    if len(resp.Choices) == 0 {
        return "", errors.New("empty response from model")
    }
    return resp.Choices[0].Message.Content, nil
}

// TextSummarize обрабатывает текстовый запрос и создает выжимку
func (c *OpenRouterClient) TextSummarize(ctx context.Context, prompt, text string, maxTokens int) (string, error) {
    if text == "" {
        return "", errors.New("text is required")
    }

    // Формируем запрос к LLM с промптом и текстом
    fullPrompt := fmt.Sprintf("%s\n\n%s", prompt, text)
    
    req := openai.ChatCompletionRequest{
        Model: c.model,
        Messages: []openai.ChatCompletionMessage{
            {
                Role: openai.ChatMessageRoleSystem,
                Content: "Ты - помощник, который создает краткие и информативные выжимки из текстов.",
            },
            {
                Role:    openai.ChatMessageRoleUser,
                Content: fullPrompt,
            },
        },
    }
    
    if maxTokens > 0 {
        req.MaxTokens = maxTokens
    }

    resp, err := c.client.CreateChatCompletion(ctx, req)
    if err != nil {
        return "", fmt.Errorf("openrouter text summarize request failed: %w", err)
    }

    if len(resp.Choices) == 0 {
        return "", errors.New("empty response from model")
    }
    
    return resp.Choices[0].Message.Content, nil
}
