package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/equaltoai/lesser-soul/pkg/config"
	"github.com/equaltoai/lesser-soul/pkg/inference"
	"github.com/equaltoai/lesser-soul/pkg/soul"
)

var inferenceClient inference.InferenceClient

func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" {
		fmt.Fprintf(os.Stdout, "%s agent-runner (lambda stub; not running locally)\n", soul.Name)
		return
	}

	stage, err := config.StageFromEnv()
	if err != nil {
		log.Printf("agent-runner: %v", err)
	} else {
		log.Printf("agent-runner: stage=%s", stage)
	}

	instanceDomain, err := config.InstanceDomainFromEnv()
	if err != nil {
		log.Printf("agent-runner: %v", err)
	} else {
		log.Printf("agent-runner: instance_domain=%s", instanceDomain)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	inf, err := initInferenceClient(ctx)
	if err != nil {
		log.Fatalf("agent-runner: %v", err)
	}
	inferenceClient = inf

	lambda.Start(handleSQSEvent)
}

func initInferenceClient(ctx context.Context) (inference.InferenceClient, error) {
	urlPath, keyPath, err := config.InferenceSSMPathsFromEnv()
	if err != nil {
		return nil, err
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)
	baseURL, apiKey, err := inference.LoadURLAndKeyFromSSM(ctx, ssmClient, urlPath, keyPath)
	if err != nil {
		return nil, err
	}
	log.Printf("agent-runner: loaded inference base_url from ssm (%s)", urlPath)

	infClient, err := inference.NewClient(baseURL, apiKey)
	if err != nil {
		return nil, err
	}
	return infClient, nil
}

func handleSQSEvent(_ context.Context, event events.SQSEvent) error {
	log.Printf("agent-runner: received %d SQS record(s)", len(event.Records))
	for _, record := range event.Records {
		log.Printf("agent-runner: message_id=%s body_len=%d", record.MessageId, len(record.Body))
	}
	return nil
}
