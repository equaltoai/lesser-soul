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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/equaltoai/lesser-soul/pkg/config"
	"github.com/equaltoai/lesser-soul/pkg/inference"
	"github.com/equaltoai/lesser-soul/pkg/runner"
	"github.com/equaltoai/lesser-soul/pkg/soul"
)

var inferenceClient inference.InferenceClient
var runnerService *runner.Service

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

	tableName, err := config.StateTableNameFromEnv()
	if err != nil {
		log.Fatalf("agent-runner: %v", err)
	}
	lesserGraphQLURL, err := config.LesserGraphQLURLFromEnv()
	if err != nil {
		log.Fatalf("agent-runner: %v", err)
	}
	resultsQueueURL, err := config.ResultsQueueURLFromEnv()
	if err != nil {
		log.Fatalf("agent-runner: %v", err)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("agent-runner: load aws config: %v", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)
	baseURL, apiKey, err := loadInferenceURLAndKey(ctx, ssmClient)
	if err != nil {
		log.Fatalf("agent-runner: %v", err)
	}
	log.Printf("agent-runner: loaded inference base_url from ssm (%s)", os.Getenv(config.EnvSoulInferenceURLSSMPath))

	inf, err := inference.NewClient(baseURL, apiKey)
	if err != nil {
		log.Fatalf("agent-runner: %v", err)
	}
	inferenceClient = inf

	db := dynamodb.NewFromConfig(awsCfg)
	sqsClient := sqs.NewFromConfig(awsCfg)
	runnerService, err = runner.NewService(tableName, instanceDomain, lesserGraphQLURL, resultsQueueURL, db, ssmClient, sqsClient, inferenceClient)
	if err != nil {
		log.Fatalf("agent-runner: %v", err)
	}

	lambda.Start(handleSQSEvent)
}

func loadInferenceURLAndKey(ctx context.Context, ssmClient *ssm.Client) (string, string, error) {
	urlPath, keyPath, err := config.InferenceSSMPathsFromEnv()
	if err != nil {
		return "", "", err
	}
	baseURL, apiKey, err := inference.LoadURLAndKeyFromSSM(ctx, ssmClient, urlPath, keyPath)
	if err != nil {
		return "", "", err
	}
	return baseURL, apiKey, nil
}

func handleSQSEvent(_ context.Context, event events.SQSEvent) error {
	log.Printf("agent-runner: received %d SQS record(s)", len(event.Records))

	bodies := make([]string, 0, len(event.Records))
	for _, record := range event.Records {
		log.Printf("agent-runner: message_id=%s body_len=%d", record.MessageId, len(record.Body))
		bodies = append(bodies, record.Body)
	}
	return runnerService.HandleSQSEvent(context.Background(), bodies)
}
