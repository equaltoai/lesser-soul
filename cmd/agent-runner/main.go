package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/equaltoai/lesser-soul/pkg/config"
	"github.com/equaltoai/lesser-soul/pkg/soul"
)

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

	lambda.Start(handleSQSEvent)
}

func handleSQSEvent(_ context.Context, event events.SQSEvent) error {
	log.Printf("agent-runner: received %d SQS record(s)", len(event.Records))
	for _, record := range event.Records {
		log.Printf("agent-runner: message_id=%s body_len=%d", record.MessageId, len(record.Body))
	}
	return nil
}
