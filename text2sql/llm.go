package text2sql

import (
	"context"
	"fmt"

	"github.com/liushuangls/go-anthropic/v2"
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/genai"
)

// LLMClient is an interface. Any struct with a Complete method can be used.
type LLMClient interface {
	Complete(ctx context.Context, systemPrompt, userMessage string) (string, error)
}

// --- OpenAI ---

type OpenAIClient struct {
	client *openai.Client
	model  string
}

func NewOpenAIClient(client *openai.Client, model string) *OpenAIClient {
	return &OpenAIClient{client: client, model: model}
}

func (o *OpenAIClient) Complete(ctx context.Context, system, user string) (string, error) {
	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       o.model,
		Temperature: 0,
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

// --- Anthropic ---

type AnthropicClient struct {
	client *anthropic.Client
	model  string
}

func NewAnthropicClient(client *anthropic.Client, model string) *AnthropicClient {
	return &AnthropicClient{client: client, model: model}
}

func (a *AnthropicClient) Complete(ctx context.Context, system, user string) (string, error) {
	resp, err := a.client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model:     anthropic.Model(a.model),
		System:    system,
		MaxTokens: 512,
		Messages: []anthropic.Message{
			{Role: anthropic.RoleUser, Content: []anthropic.MessageContent{anthropic.NewTextMessageContent(user)}},
		},
		Temperature: func(f float32) *float32 { return &f }(0.0),
	})
	if err != nil {
		return "", fmt.Errorf("anthropic: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("anthropic: no content returned")
	}

	// Text is a *string in some versions of the SDK, check and dereference
	if resp.Content[0].Text != nil {
		return *resp.Content[0].Text, nil
	}
	return "", fmt.Errorf("anthropic: content text is nil")
}

// --- Gemini ---

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(client *genai.Client, model string) *GeminiClient {
	return &GeminiClient{client: client, model: model}
}

func (g *GeminiClient) Complete(ctx context.Context, system, user string) (string, error) {
	resp, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(system+"\n\nUser Question: "+user), &genai.GenerateContentConfig{
		Temperature: func(f float32) *float32 { return &f }(0.0),
		MaxOutputTokens: 512,
	})
	if err != nil {
		return "", fmt.Errorf("gemini: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: no candidates returned")
	}

	// For genai SDK, parts are structs with a Text field
	part := resp.Candidates[0].Content.Parts[0]
	if part.Text != "" {
		return part.Text, nil
	}
	
	return "", fmt.Errorf("gemini: could not extract text from response")
}
