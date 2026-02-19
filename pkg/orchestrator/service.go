package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/equaltoai/lesser-soul/pkg/models"
)

type DynamoDBAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

type SQSAPI interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type Service struct {
	tableName          string
	instanceDomain     string
	researcherQueueURL string
	db                 DynamoDBAPI
	sqs                SQSAPI
	now                func() time.Time
}

func NewService(tableName string, instanceDomain string, researcherQueueURL string, db DynamoDBAPI, sqsClient SQSAPI) (*Service, error) {
	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		return nil, fmt.Errorf("missing tableName")
	}
	instanceDomain = strings.TrimSpace(instanceDomain)
	if instanceDomain == "" {
		return nil, fmt.Errorf("missing instanceDomain")
	}
	researcherQueueURL = strings.TrimSpace(researcherQueueURL)
	if researcherQueueURL == "" {
		return nil, fmt.Errorf("missing researcherQueueURL")
	}
	if db == nil {
		return nil, fmt.Errorf("missing dynamodb client")
	}
	if sqsClient == nil {
		return nil, fmt.Errorf("missing sqs client")
	}

	return &Service{
		tableName:          tableName,
		instanceDomain:     instanceDomain,
		researcherQueueURL: researcherQueueURL,
		db:                 db,
		sqs:                sqsClient,
		now:                time.Now,
	}, nil
}

func (s *Service) CreateTask(ctx context.Context, goal string, requestorID string) (string, string, error) {
	goal = strings.TrimSpace(goal)
	if goal == "" {
		return "", "", fmt.Errorf("missing goal")
	}

	requestorID = strings.TrimSpace(requestorID)
	if requestorID == "" {
		requestorID = "unknown"
	}

	now := s.now().UTC()
	createdAt := now.Format(time.RFC3339)
	ttl := models.TTL30Days(now)

	taskID := models.NewTaskID()
	subTaskSK := models.NewSubTaskSK()

	if _, err := s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item: map[string]dynamotypes.AttributeValue{
			"pk":              &dynamotypes.AttributeValueMemberS{Value: taskID},
			"sk":              &dynamotypes.AttributeValueMemberS{Value: "META"},
			"instance_domain": &dynamotypes.AttributeValueMemberS{Value: s.instanceDomain},
			"status":          &dynamotypes.AttributeValueMemberS{Value: string(models.TaskStatusRunning)},
			"created_at":      &dynamotypes.AttributeValueMemberS{Value: createdAt},
			"goal":            &dynamotypes.AttributeValueMemberS{Value: goal},
			"requestor_id":    &dynamotypes.AttributeValueMemberS{Value: requestorID},
			"total_subtasks":  &dynamotypes.AttributeValueMemberN{Value: "1"},
			"done_subtasks":   &dynamotypes.AttributeValueMemberN{Value: "0"},
			"failed_subtasks": &dynamotypes.AttributeValueMemberN{Value: "0"},
			"ttl":             &dynamotypes.AttributeValueMemberN{Value: strconv.FormatInt(ttl, 10)},
		},
	}); err != nil {
		return "", "", fmt.Errorf("put task: %w", err)
	}

	if _, err := s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item: map[string]dynamotypes.AttributeValue{
			"pk":           &dynamotypes.AttributeValueMemberS{Value: taskID},
			"sk":           &dynamotypes.AttributeValueMemberS{Value: subTaskSK},
			"agent_type":   &dynamotypes.AttributeValueMemberS{Value: string(models.AgentTypeResearcher)},
			"status":       &dynamotypes.AttributeValueMemberS{Value: string(models.SubTaskStatusQueued)},
			"goal":         &dynamotypes.AttributeValueMemberS{Value: goal},
			"queue_url":    &dynamotypes.AttributeValueMemberS{Value: s.researcherQueueURL},
			"tokens_in":    &dynamotypes.AttributeValueMemberN{Value: "0"},
			"tokens_out":   &dynamotypes.AttributeValueMemberN{Value: "0"},
			"ttl":          &dynamotypes.AttributeValueMemberN{Value: strconv.FormatInt(ttl, 10)},
			"started_at":   &dynamotypes.AttributeValueMemberS{Value: ""},
			"completed_at": &dynamotypes.AttributeValueMemberS{Value: ""},
		},
	}); err != nil {
		return "", "", fmt.Errorf("put subtask: %w", err)
	}

	body, err := json.Marshal(models.SubTaskQueueMessage{
		TaskID:    taskID,
		SubTaskSK: subTaskSK,
		AgentType: models.AgentTypeResearcher,
	})
	if err != nil {
		return "", "", fmt.Errorf("marshal queue message: %w", err)
	}

	if _, err := s.sqs.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.researcherQueueURL),
		MessageBody: aws.String(string(body)),
	}); err != nil {
		return "", "", fmt.Errorf("send researcher message: %w", err)
	}

	return taskID, subTaskSK, nil
}

func (s *Service) ApplyResult(ctx context.Context, msg models.SubTaskResultMessage) error {
	taskID := strings.TrimSpace(msg.TaskID)
	subTaskSK := strings.TrimSpace(msg.SubTaskSK)
	if taskID == "" || subTaskSK == "" {
		return fmt.Errorf("result missing task_id or subtask_sk")
	}

	now := s.now().UTC().Format(time.RFC3339)

	status := models.SubTaskStatusDone
	if strings.TrimSpace(msg.Error) != "" {
		status = models.SubTaskStatusFailed
	}
	if status == models.SubTaskStatusDone && strings.TrimSpace(msg.LesserNoteID) == "" {
		return fmt.Errorf("result missing lesser_note_id")
	}

	updateSubtaskExpr := "SET #status = :status, completed_at = :completed_at, tokens_in = :tokens_in, tokens_out = :tokens_out"
	values := map[string]dynamotypes.AttributeValue{
		":status":       &dynamotypes.AttributeValueMemberS{Value: string(status)},
		":completed_at": &dynamotypes.AttributeValueMemberS{Value: now},
		":tokens_in":    &dynamotypes.AttributeValueMemberN{Value: strconv.Itoa(msg.TokensIn)},
		":tokens_out":   &dynamotypes.AttributeValueMemberN{Value: strconv.Itoa(msg.TokensOut)},
	}
	if status == models.SubTaskStatusDone {
		updateSubtaskExpr += ", lesser_note_id = :lesser_note_id"
		values[":lesser_note_id"] = &dynamotypes.AttributeValueMemberS{Value: strings.TrimSpace(msg.LesserNoteID)}
	} else {
		updateSubtaskExpr += ", error = :error"
		values[":error"] = &dynamotypes.AttributeValueMemberS{Value: strings.TrimSpace(msg.Error)}
	}

	if _, err := s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]dynamotypes.AttributeValue{
			"pk": &dynamotypes.AttributeValueMemberS{Value: taskID},
			"sk": &dynamotypes.AttributeValueMemberS{Value: subTaskSK},
		},
		UpdateExpression:          aws.String(updateSubtaskExpr),
		ExpressionAttributeNames:  map[string]string{"#status": "status"},
		ExpressionAttributeValues: values,
	}); err != nil {
		return fmt.Errorf("update subtask: %w", err)
	}

	taskStatus := models.TaskStatusDone
	if status == models.SubTaskStatusFailed {
		taskStatus = models.TaskStatusFailed
	}

	if _, err := s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]dynamotypes.AttributeValue{
			"pk": &dynamotypes.AttributeValueMemberS{Value: taskID},
			"sk": &dynamotypes.AttributeValueMemberS{Value: "META"},
		},
		UpdateExpression:         aws.String("SET #status = :status, done_subtasks = :done_subtasks, failed_subtasks = :failed_subtasks"),
		ExpressionAttributeNames: map[string]string{"#status": "status"},
		ExpressionAttributeValues: map[string]dynamotypes.AttributeValue{
			":status":        &dynamotypes.AttributeValueMemberS{Value: string(taskStatus)},
			":done_subtasks": &dynamotypes.AttributeValueMemberN{Value: "1"},
			":failed_subtasks": &dynamotypes.AttributeValueMemberN{Value: func() string {
				if status == models.SubTaskStatusFailed {
					return "1"
				}
				return "0"
			}()},
		},
	}); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	return nil
}
