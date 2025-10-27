package internal

import (
	"encoding/json"
	"os"

	"github.com/pkoukk/tiktoken-go"
)

type Token struct {
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`
}

var tokens = new([]Token)
var tokenEncoder = new(tiktoken.Tiktoken)

func InitTokens() {
	loadTokens(AppConfig.Common.TokensFile)
	t, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		panic(err)
	}
	tokenEncoder = t
}

func loadTokens(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	json.Unmarshal(content, tokens)
	return nil
}

func CountTokens(content string) int {
	ids := tokenEncoder.Encode(content, nil, nil)
	return len(ids)
}

func CountTokensFast(content string) int {
	return int(float64(len(content)) / 3.75)
}

func GetTokens(offset int, count int) *[]Token {

	offset = offset % len(*tokens)

	tokenCount := 0
	result := make([]Token, 0)
	idx := 0
	for tokenCount < count {
		pick := (offset + idx) % len(*tokens)
		token := (*tokens)[pick]
		tokenCount += token.Tokens
		result = append(result, token)
		idx++
	}
	if tokenCount > count {
		result = result[:len(result)-1]
	}
	return &result
}
