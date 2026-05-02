package text2sql

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// LLMClient is an interface. Any struct with a Complete method can be used.
// This makes it easy to switch providers or write tests.
type LLMClient interface {
	Complete(ctx context.Context, systemPrompt, userMessage string) (string, error)
}

// OpenAIClient implements LLMClient using the OpenAI API
type OpenAIClient struct {
	client *openai.Client
	model  string // e.g. "gpt-4o"
}

func NewOpenAIClient(client *openai.Client, model string) *OpenAIClient {
	return &OpenAIClient{client: client, model: model}
}

func (o *OpenAIClient) Complete(ctx context.Context, system, user string) (string, error) {
	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       o.model,
		Temperature: 0, // IMPORTANT: always 0 for deterministic SQL output
		MaxTokens:   512,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	})
	if err != nil {
		return "", fmt.Errorf("openai: %w", err)
	}
	return resp.Choices[0].Message.Content, nil
}
