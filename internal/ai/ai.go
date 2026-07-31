package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

var exampleStyles = []struct {
	Name        string
	Description string
}{
	{Name: "Casual", Description: "Casual (casual speech between friends)"},
	{Name: "Polite", Description: "Polite (desu/masu form)"},
	{Name: "Formal", Description: "Formal (honorific speech)"},
	{Name: "Written", Description: "Written/Literary (written Japanese style)"},
	{Name: "Question", Description: "Question form (interrogative sentence)"},
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

	return parseExamplesResponse(text)
}

func (c *Client) buildGenerationConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		Temperature: &c.config.Temperature,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: "あなたは日本語教師です。日本語学習者のために例文を生成します。"},
			},
		},
	}
}

func buildExamplePrompt(word string) string {
	var styleLines []string
	for i, s := range exampleStyles {
		styleLines = append(styleLines, fmt.Sprintf("%d. %s", i+1, s.Description))
	}

	return fmt.Sprintf(
		`「%s」という単語を使った例文を5つ生成してください。それぞれ異なるスタイルにしてください。

スタイル:
%s

出力形式（各行を正確にこの形式にしてください）:
1. [スタイル名]: [日本語の例文]
2. [スタイル名]: [日本語の例文]`,
		word,
		strings.Join(styleLines, "\n"),
	)
}

func parseExamplesResponse(text string) ([]GeneratedExample, error) {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var examples []GeneratedExample

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		content := line
		if idx := strings.Index(line, ". "); idx != -1 && idx < 3 {
			content = line[idx+2:]
		}

		parts := strings.SplitN(content, ": ", 2)
		if len(parts) != 2 {
			continue
		}

		examples = append(examples, GeneratedExample{
			Style: strings.TrimSpace(parts[0]),
			Text:  strings.TrimSpace(parts[1]),
		})
	}

	if len(examples) == 0 {
		return nil, errors.New("no valid examples found in gemini response")
	}

	return examples, nil
}
