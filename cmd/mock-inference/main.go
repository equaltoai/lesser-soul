package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/equaltoai/lesser-soul/pkg/soul"
)

func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		fmt.Fprintf(os.Stdout, "%s mock-inference (lambda stub; not running locally)\n", soul.Name)
		return
	}

	lambda.Start(handleFunctionURL)
}

func handleFunctionURL(_ context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	method := strings.ToUpper(req.RequestContext.HTTP.Method)

	path := req.RawPath
	if path == "" {
		path = "/"
	}
	cleanPath := strings.TrimRight(path, "/")
	if cleanPath == "" {
		cleanPath = "/"
	}

	if method == http.MethodOptions {
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusNoContent,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "authorization,content-type",
				"Access-Control-Allow-Methods": "POST,OPTIONS",
			},
		}, nil
	}

	if method != http.MethodPost || (cleanPath != "/chat/completions" && cleanPath != "/v1/chat/completions") {
		return jsonResponse(http.StatusNotFound, map[string]any{"error": "not_found"}), nil
	}

	var payload struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	if req.Body != "" {
		if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
			return jsonResponse(http.StatusBadRequest, map[string]any{"error": "invalid_json"}), nil
		}
	}

	userText := ""
	for _, m := range payload.Messages {
		if strings.ToLower(strings.TrimSpace(m.Role)) == "user" {
			userText = strings.TrimSpace(m.Content)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	content := fmt.Sprintf("mock inference (%s): %s", now, userText)

	log.Printf("mock-inference: model=%q user_len=%d", payload.Model, len(userText))

	return jsonResponse(http.StatusOK, map[string]any{
		"choices": []any{
			map[string]any{
				"message": map[string]any{
					"role":    "assistant",
					"content": content,
				},
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     10,
			"completion_tokens": 20,
			"total_tokens":      30,
		},
	}), nil
}

func jsonResponse(statusCode int, body any) events.LambdaFunctionURLResponse {
	b, err := json.Marshal(body)
	if err != nil {
		b = []byte(`{"error":"internal_error"}`)
		statusCode = http.StatusInternalServerError
	}
	return events.LambdaFunctionURLResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(b),
	}
}
