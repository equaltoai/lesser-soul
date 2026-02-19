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
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/equaltoai/lesser-soul/pkg/config"
	"github.com/equaltoai/lesser-soul/pkg/models"
	"github.com/equaltoai/lesser-soul/pkg/orchestrator"
	"github.com/equaltoai/lesser-soul/pkg/soul"
)

var svc *orchestrator.Service

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

	tableName, err := config.StateTableNameFromEnv()
	if err != nil {
		log.Fatalf("orchestrator: %v", err)
	}

	researcherQueueURL, err := config.ResearcherQueueURLFromEnv()
	if err != nil {
		log.Fatalf("orchestrator: %v", err)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("orchestrator: load aws config: %v", err)
	}

	db := dynamodb.NewFromConfig(awsCfg)
	sqsClient := sqs.NewFromConfig(awsCfg)

	svc, err = orchestrator.NewService(tableName, instanceDomain, researcherQueueURL, db, sqsClient)
	if err != nil {
		log.Fatalf("orchestrator: %v", err)
	}

	lambda.Start(handleEvent)
}

func handleEvent(ctx context.Context, raw json.RawMessage) (any, error) {
	var envelope struct {
		Records json.RawMessage `json:"Records"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Records) > 0 {
		var event events.SQSEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil, err
		}
		return nil, handleResultsSQSEvent(ctx, event)
	}

	var req events.LambdaFunctionURLRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}
	resp, err := handleFunctionURLEvent(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func handleFunctionURLEvent(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
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
			if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
				return jsonResponse(http.StatusBadRequest, map[string]any{
					"error": "invalid_json",
				}, nil), nil
			}
		}
		payload.Goal = strings.TrimSpace(payload.Goal)
		if payload.Goal == "" {
			return jsonResponse(http.StatusBadRequest, map[string]any{
				"error": "missing_goal",
			}, nil), nil
		}

		taskID, subTaskSK, err := svc.CreateTask(ctx, payload.Goal, "unknown")
		if err != nil {
			log.Printf("orchestrator: create task: %v", err)
			return jsonResponse(http.StatusInternalServerError, map[string]any{
				"error": "internal_error",
			}, nil), nil
		}

		return jsonResponse(http.StatusOK, map[string]any{
			"task_id":    taskID,
			"subtask_sk": subTaskSK,
			"status":     "RUNNING",
		}, nil), nil
	}

	return jsonResponse(http.StatusNotFound, map[string]any{
		"error": "not_found",
	}, nil), nil
}

func handleResultsSQSEvent(ctx context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		var msg models.SubTaskResultMessage
		if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
			return fmt.Errorf("decode result message: %w", err)
		}
		if err := svc.ApplyResult(ctx, msg); err != nil {
			return err
		}
	}
	return nil
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
