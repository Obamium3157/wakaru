package ai

import "context"

type ExampleGenerator interface {
	GenerateExamples(ctx context.Context, word string) ([]GeneratedExample, error)
}

type GeneratedExample struct {
	Style string `json:"style"`
	Text  string `json:"text"`
}
