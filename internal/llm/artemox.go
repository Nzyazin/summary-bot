package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ArtemoxClient реализует интерфейс Client для работы с API Artemox
type ArtemoxClient struct {
	apiKey   string
	baseURL  string
	model    string
	client   *http.Client
}

// Структуры для запросов и ответов Artemox API
type artemoxMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type artemoxRequest struct {
	Model    string           `json:"model"`
	Messages []artemoxMessage `json:"messages"`
}

type artemoxChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type artemoxResponse struct {
	Choices []artemoxChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewArtemoxClient создает новый клиент для работы с API Artemox
func NewArtemoxClient(apiKey, model string) (*ArtemoxClient, error) {
	if apiKey == "" {
		return nil, errors.New("artemox api key is empty")
	}

	baseURL := "https://api.artemox.com/v1"
	
	if model == "" {
		model = "gpt-4o-mini" // Модель по умолчанию
	}

	return &ArtemoxClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}, nil
}

func (c *ArtemoxClient) TextSummarize(ctx context.Context, prompt, text string, maxTokens int) (string, error) {
	if text == "" {
		return "", errors.New("text is required")
	}

	systemPrompt := "Ты - помощник, который создает краткие и информативные выжимки из текстов."
	userPrompt := fmt.Sprintf("%s\n\n%s", prompt, text)

	messages := []artemoxMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	return c.sendRequest(ctx, messages, maxTokens)
}

func (c *ArtemoxClient) VisionDescribe(ctx context.Context, prompt, imageURL string, maxTokens int) (string, error) {
	if imageURL == "" {
		return "", errors.New("image url is required")
	}

	// Для работы с изображениями может потребоваться другой формат запроса
	// В данном примере предполагаем, что API поддерживает URL изображений в тексте
	content := fmt.Sprintf("%s\nImage URL: %s", prompt, imageURL)
	
	messages := []artemoxMessage{
		{Role: "user", Content: content},
	}

	return c.sendRequest(ctx, messages, maxTokens)
}

// sendRequest отправляет запрос к API Artemox
func (c *ArtemoxClient) sendRequest(ctx context.Context, messages []artemoxMessage, maxTokens int) (string, error) {
	reqBody := artemoxRequest{
		Model:    c.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	var artemoxResp artemoxResponse
	if err := json.Unmarshal(body, &artemoxResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if artemoxResp.Error != nil && artemoxResp.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", artemoxResp.Error.Message)
	}

	if len(artemoxResp.Choices) == 0 {
		return "", errors.New("empty response from model")
	}

	return artemoxResp.Choices[0].Message.Content, nil
}
