package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/equaltoai/lesser-soul/pkg/config"
	"github.com/equaltoai/lesser-soul/pkg/soul"
)

func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		fmt.Fprintf(os.Stdout, "%s orchestrator (lambda stub; not running locally)\n", soul.Name)
		return
	}

	stage, err := config.StageFromEnv()
	if err != nil {
		log.Printf("orchestrator: %v", err)
	} else {
		log.Printf("orchestrator: stage=%s", stage)
	}

	instanceDomain, err := config.InstanceDomainFromEnv()
	if err != nil {
		log.Printf("orchestrator: %v", err)
	} else {
		log.Printf("orchestrator: instance_domain=%s", instanceDomain)
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

	auth := headerValue(req.Headers, "Authorization")
	if !strings.HasPrefix(auth, "Bearer ") || strings.TrimSpace(strings.TrimPrefix(auth, "Bearer ")) == "" {
		return jsonResponse(http.StatusUnauthorized, map[string]any{
			"error": "missing_or_invalid_authorization",
		}, map[string]string{
			"WWW-Authenticate": "Bearer",
		}), nil
	}

	if method == http.MethodPost && cleanPath == "/soul/tasks" {
		var payload struct {
			Goal string `json:"goal"`
		}
		if req.Body != "" {
			_ = json.Unmarshal([]byte(req.Body), &payload)
		}

		return jsonResponse(http.StatusOK, map[string]any{
			"task_id": "stub",
			"status":  "STUB",
			"goal":    payload.Goal,
		}, nil), nil
	}

	return jsonResponse(http.StatusNotFound, map[string]any{
		"error": "not_found",
	}, nil), nil
}

func headerValue(headers map[string]string, name string) string {
	for k, v := range headers {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	return ""
}

func jsonResponse(statusCode int, body any, extraHeaders map[string]string) events.LambdaFunctionURLResponse {
	b, err := json.Marshal(body)
	if err != nil {
		b = []byte(`{"error":"internal_error"}`)
		statusCode = http.StatusInternalServerError
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	for k, v := range extraHeaders {
		headers[k] = v
	}

	return events.LambdaFunctionURLResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(b),
	}
}
