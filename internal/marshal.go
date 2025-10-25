package internal

import (
	"time"

	"github.com/bytedance/sonic"
)

var respEncoder = sonic.ConfigFastest

func MarshalChunk(id string, created time.Time, model string, content string,
	promptTokens, completionTokens, totalTokens int) (*[]byte, error) {

	ch := Chunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: created.Unix(),
		Model:   model,
		Choices: []Choice{{
			Index: 0,
			Delta: Delta{
				Content:          content,
				ReasoningContent: nil,
				Role:             "assistant",
			},
			FinishReason: nil,
		}},
		Usage: Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
	}

	b, err := respEncoder.Marshal(&ch)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func MarshalCompletion(id string, created time.Time, model string, content string,
	promptTokens, completionTokens, totalTokens int) (*[]byte, error) {

	ch := Completion{
		ID:      id,
		Object:  "chat.completion",
		Created: created.Unix(),
		Model:   model,
		Choices: []CompletionChoice{{
			Index: 0,
			Message: Message{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: nil,
		}},
		Usage: Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
	}

	b, err := respEncoder.Marshal(&ch)
	if err != nil {
		return nil, err
	}

	return &b, nil
}
