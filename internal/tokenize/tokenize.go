// Package tokenize provides API for splitting Japanese sentences into parts
package tokenize

import (
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

type TokenizeResult struct {
	Token    string
	BaseForm string
}

func Tokenize(sentence string) ([]TokenizeResult, error) {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, err
	}

	tokens := t.Tokenize(sentence)

	var result []TokenizeResult

	for _, token := range tokens {
		baseForm, _ := token.BaseForm()

		result = append(result, TokenizeResult{Token: token.Surface, BaseForm: baseForm})
	}

	return result, nil
}
