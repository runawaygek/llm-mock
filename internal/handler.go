package internal

import (
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var requestDecoder = sonic.ConfigFastest

type MockRequest struct {
	ReqID        string
	Start        time.Time
	modelConfig  ModelConfig
	PromptTokens int
	OutputTokens int
	Choices      *[]Token
}

var (
	HandlerEnterCount  atomic.Uint64
	HandlerActiveCount atomic.Int64
)

func ModelsHandler(c *gin.Context) {
	models := make([]string, 0, len(AppConfig.ModelMap))
	for name := range AppConfig.ModelMap {
		models = append(models, name)
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func ChatHandler(c *gin.Context) {

	if AppConfig.Server.PprofEnabled {
		HandlerEnterCount.Add(1)
		HandlerActiveCount.Add(1)
		defer HandlerActiveCount.Add(-1)
	}

	req_id := uuid.New().String()
	start := time.Now()
	var request ChatRequest
	err := requestDecoder.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	modelConfig, ok := AppConfig.ModelMap[request.Model]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model not found"})
		return
	}
	maxTokens := min(request.MaxTokens, modelConfig.MaxOutputTokens)

	prompt := strings.Builder{}
	for _, message := range request.Messages {
		prompt.Write(message.Content)
	}

	promptTokens := CountTokensFast(prompt.String())
	if promptTokens > modelConfig.MaxContextTokens {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt tokens exceed max context tokens"})
		return
	}

	maxTokens = min(maxTokens, modelConfig.MaxContextTokens-promptTokens)

	if maxTokens <= 0 {
		maxTokens = 50
	}

	choices := GetTokens(promptTokens, maxTokens)

	mockRequest := MockRequest{
		ReqID:        req_id,
		Start:        start,
		modelConfig:  modelConfig,
		PromptTokens: promptTokens,
		OutputTokens: maxTokens,
		Choices:      choices,
	}

	// Logger.Info("PreProcess time cost", zap.Duration("cost", time.Since(start)))

	if request.Stream {
		handleStreamChat(c, &mockRequest)
	} else {
		handleNormalChat(c, &mockRequest)
	}
}

func handleNormalChat(c *gin.Context, mockRequest *MockRequest) {
	ttft := mockRequest.modelConfig.TTFT.GetTTFT()
	time.Sleep(time.Duration(ttft) * time.Millisecond)

	tpot_ms := 1000 / (mockRequest.modelConfig.OTPS + 1)
	tpot := time.Duration(tpot_ms) * time.Millisecond

	content := strings.Builder{}
	first := true
	for _, choice := range *mockRequest.Choices {
		if !first {
			time.Sleep(tpot)
		}
		first = false
		content.WriteString(choice.Content)
	}

	b, err := MarshalCompletion(mockRequest.ReqID, time.Now(), mockRequest.modelConfig.Name, content.String(),
		mockRequest.PromptTokens, mockRequest.OutputTokens, mockRequest.PromptTokens+mockRequest.OutputTokens)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", *b)
}

func handleStreamChat(c *gin.Context, mockRequest *MockRequest) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	ttft := mockRequest.modelConfig.TTFT.GetTTFT()
	time.Sleep(time.Duration(ttft) * time.Millisecond)

	tpot_ms := 1000 / (mockRequest.modelConfig.OTPS + 1)
	tpot := time.Duration(tpot_ms) * time.Millisecond

	first := true
	completionTokens := 0
	for _, choice := range *mockRequest.Choices {
		if !first {
			time.Sleep(tpot)
		}
		first = false
		completionTokens += choice.Tokens
		chunk, err := MarshalChunk(
			mockRequest.ReqID,
			time.Now(),
			mockRequest.modelConfig.Name,
			choice.Content,
			mockRequest.PromptTokens,
			completionTokens,
			completionTokens+mockRequest.PromptTokens,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Writer.Write([]byte("data: "))
		c.Writer.Write(*chunk)
		c.Writer.Write([]byte("\n\n"))
		c.Writer.Flush()
	}

	lastChunk, err := MarshalChunk(
		mockRequest.ReqID,
		time.Now(),
		mockRequest.modelConfig.Name,
		"",
		mockRequest.PromptTokens,
		completionTokens,
		completionTokens+mockRequest.PromptTokens,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Write([]byte("data: "))
	c.Writer.Write(*lastChunk)
	c.Writer.Write([]byte("\n\n"))

	c.Writer.Write([]byte("data: [DONE]\n\n"))
	c.Writer.Flush()
}
