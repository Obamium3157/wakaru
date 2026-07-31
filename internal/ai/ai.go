// Package ai provides AI-powered generation of Japanese example sentences.
package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

type ClientConfig struct {
	Model       string
	Temperature float32
}

type Client struct {
	aiClient *genai.Client
	config   ClientConfig
}

func New(ctx context.Context, conf ClientConfig, apiKey string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		aiClient: client,
		config:   conf,
	}, nil
}

func (c *Client) GenerateExamples(ctx context.Context, word string) ([]GeneratedExample, error) {
	if word == "" {
		return nil, errors.New("word cannot be empty")
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	config := c.buildGenerationConfig()
	prompt := buildExamplePrompt(word)

	result, err := c.aiClient.Models.GenerateContent(
		ctx,
		c.config.Model,
		genai.Text(prompt),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}

	text := result.Text()
	if text == "" {
		return nil, errors.New("gemini returned empty response")
	}

	var examples []GeneratedExample
	if err := json.Unmarshal([]byte(text), &examples); err != nil {
		return nil, fmt.Errorf("invalid JSON from gemini: %w", err)
	}

	if err := validateExamples(examples); err != nil {
		return nil, err
	}

	return examples, nil
}

func (c *Client) buildGenerationConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		Temperature:      &c.config.Temperature,
		ResponseMIMEType: "application/json",
		ResponseSchema:   buildResponseSchema(),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: systemInstruction},
			},
		},
	}
}

const systemInstruction = `あなたは日本語教師です。日本語学習者のために、指定された単語を使った自然な日本語の例文を生成します。

出力ルール:
- 要求されたJSON形式のみを出力してください。JSON以外のテキスト（マークダウン、説明、前置き、翻訳、コードブロック）は絶対に出力しないでください。
- 各例文は自然な日本語で、必ず1文にしてください。
- 同じ例文を繰り返さないでください。`
