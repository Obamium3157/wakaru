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

func Tokenize(sentence string) ([]tokenizer.Token, error) {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, err
	}

	tokens := t.Tokenize(sentence)

	return tokens, nil
}
