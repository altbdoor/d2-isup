package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

const MAX_ATTEMPTS = 3

func ParseGemini(prompt, rssXml string) ([]byte, error) {
	var err error

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("failed to get GOOGLE_API_KEY env")
	}

	client := openai.NewClient(
		option.WithBaseURL("https://generativelanguage.googleapis.com/v1beta/openai/"),
		option.WithAPIKey(apiKey),
		option.WithMaxRetries(0),
	)

	ctx := context.Background()
	var genResp *openai.ChatCompletion
	attempts := 0

	now := time.Now()
	fixedPrompt := strings.ReplaceAll(prompt, "__CURRENT_DATE__", now.Format(time.RFC3339))

	for attempts < MAX_ATTEMPTS {
		genResp, err = client.Chat.Completions.New(
			ctx,
			openai.ChatCompletionNewParams{
				Model:       shared.ChatModel("gemini-2.0-flash-lite"),
				Temperature: openai.Float(0.6),
				MaxTokens:   openai.Int(8000),
				N:           openai.Int(1),
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.SystemMessage(fixedPrompt),
					openai.UserMessage("```xml\n" + rssXml + "\n```"),
				},
				ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
					OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
				},
			},
			option.WithRequestTimeout(10*time.Second),
		)

		if err == nil {
			break
		}

		attempts++

		attemptLog := fmt.Sprintf("attempt %d failed to generate \n%v", attempts, err)
		slog.Warn(attemptLog)

		slog.Warn("retrying in 2s...")
		time.Sleep(2 * time.Second)
	}

	if attempts == MAX_ATTEMPTS {
		return nil, fmt.Errorf("max retries reached")
	}

	aiResponse := genResp.Choices[0].Message.Content

	maintenanceData := []MaintenanceData{}
	err = json.Unmarshal([]byte(aiResponse), &maintenanceData)

	if err != nil {
		return nil, fmt.Errorf("failed to json parse: %w \n\n%v", err, aiResponse)
	}

	outputBytes, err := json.MarshalIndent(maintenanceData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to json marshal: %w \n\n%v", err, aiResponse)
	}

	return outputBytes, nil
}
