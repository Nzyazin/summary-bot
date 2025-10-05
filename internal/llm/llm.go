package llm

import "context"

type Client interface {
    // VisionDescribe обрабатывает запрос с изображением
    VisionDescribe(ctx context.Context, prompt, imageURL string, maxTokens int) (string, error)
    
    // TextSummarize обрабатывает текстовый запрос и создает выжимку
    TextSummarize(ctx context.Context, prompt, text string, maxTokens int) (string, error)
}
