package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/equaltoai/lesser-soul/pkg/inference"
	"github.com/equaltoai/lesser-soul/pkg/lesser"
	"github.com/equaltoai/lesser-soul/pkg/models"
	"github.com/oklog/ulid/v2"
)

type DynamoDBAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

type SSMAPI interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

type SQSAPI interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type Service struct {
	tableName      string
	instanceDomain string
	graphQLURL     string
	resultsQueue   string

	db  DynamoDBAPI
	ssm SSMAPI
	sqs SQSAPI

	inference inference.InferenceClient

	now func() time.Time
}

func NewService(tableName string, instanceDomain string, graphQLURL string, resultsQueueURL string, db DynamoDBAPI, ssmClient SSMAPI, sqsClient SQSAPI, inf inference.InferenceClient) (*Service, error) {
	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		return nil, fmt.Errorf("missing tableName")
	}
	instanceDomain = strings.TrimSpace(instanceDomain)
	if instanceDomain == "" {
		return nil, fmt.Errorf("missing instanceDomain")
	}
	graphQLURL = strings.TrimSpace(graphQLURL)
	if graphQLURL == "" {
		return nil, fmt.Errorf("missing graphQLURL")
	}
	resultsQueueURL = strings.TrimSpace(resultsQueueURL)
	if resultsQueueURL == "" {
		return nil, fmt.Errorf("missing resultsQueueURL")
	}
	if db == nil {
		return nil, fmt.Errorf("missing dynamodb client")
	}
	if ssmClient == nil {
		return nil, fmt.Errorf("missing ssm client")
	}
	if sqsClient == nil {
		return nil, fmt.Errorf("missing sqs client")
	}
	if inf == nil {
		return nil, fmt.Errorf("missing inference client")
	}

	return &Service{
		tableName:      tableName,
		instanceDomain: instanceDomain,
		graphQLURL:     graphQLURL,
		resultsQueue:   resultsQueueURL,
		db:             db,
		ssm:            ssmClient,
		sqs:            sqsClient,
		inference:      inf,
		now:            time.Now,
	}, nil
}

func (s *Service) HandleSQSEvent(ctx context.Context, eventRecords []string) error {
	for _, body := range eventRecords {
		if err := s.handleMessage(ctx, body); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) handleMessage(ctx context.Context, body string) error {
	var msg models.SubTaskQueueMessage
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		return fmt.Errorf("decode message: %w", err)
	}

	taskID := strings.TrimSpace(msg.TaskID)
	subTaskSK := strings.TrimSpace(msg.SubTaskSK)
	if taskID == "" || subTaskSK == "" {
		return fmt.Errorf("message missing task_id or subtask_sk")
	}

	agentType := models.AgentType(strings.TrimSpace(string(msg.AgentType)))
	if agentType == "" {
		return fmt.Errorf("message missing agent_type")
	}

	return s.runAgentTurn(ctx, taskID, subTaskSK, agentType)
}

func (s *Service) runAgentTurn(ctx context.Context, taskID string, subTaskSK string, agentType models.AgentType) error {
	goal, err := s.loadSubTaskGoal(ctx, taskID, subTaskSK)
	if err != nil {
		return err
	}

	startedAt := s.now().UTC().Format(time.RFC3339)
	if err := s.markSubTaskRunning(ctx, taskID, subTaskSK, startedAt); err != nil {
		if isConditionalCheckFailed(err) {
			log.Printf("agent-runner: subtask already claimed (task_id=%s subtask_sk=%s)", taskID, subTaskSK)
			return nil
		}
		return err
	}

	agentCfg, err := s.loadAgentConfig(ctx, agentType)
	if err != nil {
		_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "ERROR", 0, 0, "", "load_agent_config_failed: "+err.Error())
		return s.publishFailure(ctx, taskID, subTaskSK, err)
	}
	if !agentCfg.Enabled {
		_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "ERROR", 0, 0, "", "agent_disabled")
		return s.publishFailure(ctx, taskID, subTaskSK, fmt.Errorf("agent %s disabled", agentType))
	}

	tokenPath := strings.TrimSpace(agentCfg.TokenSSMPath)
	if tokenPath == "" {
		return s.publishFailure(ctx, taskID, subTaskSK, fmt.Errorf("agent config missing token_ssm_path (agent_type=%s)", agentType))
	}
	agentToken, err := s.loadSecureString(ctx, tokenPath)
	if err != nil {
		return s.publishFailure(ctx, taskID, subTaskSK, fmt.Errorf("load agent token: %w", err))
	}

	lesserClient, err := lesser.NewClient(s.graphQLURL, lesser.WithBearerToken(agentToken))
	if err != nil {
		return s.publishFailure(ctx, taskID, subTaskSK, err)
	}

	mem, err := lesserClient.AgentMemorySearch(ctx, lesser.AgentMemorySearchParams{Query: goal})
	if err != nil {
		_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "ERROR", 0, 0, "", "memory_search_failed: "+err.Error())
		return s.publishFailure(ctx, taskID, subTaskSK, err)
	}

	systemPrompt, err := renderSystemPrompt(agentCfg.SystemPromptTemplate, systemPromptData{
		Goal:      goal,
		AgentType: string(agentType),
		Memory:    normalizeMemoryNotes(mem),
	})
	if err != nil {
		_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "ERROR", 0, 0, "", "render_system_prompt_failed: "+err.Error())
		return s.publishFailure(ctx, taskID, subTaskSK, err)
	}

	modelID := strings.TrimSpace(agentCfg.ModelID)
	if modelID == "" {
		modelID = "llama-3.3-70b"
	}
	maxTokens := agentCfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1200
	}

	resp, err := s.inference.Complete(ctx, inference.CompletionRequest{
		Model:        modelID,
		SystemPrompt: systemPrompt,
		Messages: []inference.Message{
			{Role: "user", Content: goal},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.2,
	})
	if err != nil {
		_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "ERROR", 0, 0, "", "inference_failed: "+err.Error())
		return s.publishFailure(ctx, taskID, subTaskSK, err)
	}
	_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "LLM_CALLED", resp.Usage.PromptTokens, resp.Usage.CompletionTokens, "", "")

	note, err := lesserClient.CreateNote(ctx, lesser.CreateNoteInput{
		Content:    resp.Content,
		Visibility: lesser.VisibilityUnlisted,
	})
	if err != nil {
		_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "ERROR", 0, 0, "", "create_note_failed: "+err.Error())
		return s.publishFailure(ctx, taskID, subTaskSK, err)
	}
	_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "NOTE_POSTED", 0, 0, note.ID, "")

	if err := s.publishResult(ctx, models.SubTaskResultMessage{
		TaskID:       taskID,
		SubTaskSK:    subTaskSK,
		LesserNoteID: note.ID,
		TokensIn:     resp.Usage.PromptTokens,
		TokensOut:    resp.Usage.CompletionTokens,
	}); err != nil {
		_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "ERROR", 0, 0, "", "publish_result_failed: "+err.Error())
		return err
	}
	_ = s.writeRunLog(ctx, taskID, subTaskSK, agentType, "RESULT_PUBLISHED", 0, 0, note.ID, "")
	return nil
}

func (s *Service) loadSubTaskGoal(ctx context.Context, taskID string, subTaskSK string) (string, error) {
	out, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.tableName),
		ConsistentRead: aws.Bool(true),
		Key: map[string]dynamotypes.AttributeValue{
			"pk": &dynamotypes.AttributeValueMemberS{Value: taskID},
			"sk": &dynamotypes.AttributeValueMemberS{Value: subTaskSK},
		},
	})
	if err != nil {
		return "", fmt.Errorf("get subtask: %w", err)
	}
	if len(out.Item) == 0 {
		return "", fmt.Errorf("subtask not found (task_id=%s subtask_sk=%s)", taskID, subTaskSK)
	}
	rawGoal, ok := out.Item["goal"].(*dynamotypes.AttributeValueMemberS)
	if !ok || strings.TrimSpace(rawGoal.Value) == "" {
		return "", fmt.Errorf("subtask missing goal (task_id=%s subtask_sk=%s)", taskID, subTaskSK)
	}
	return strings.TrimSpace(rawGoal.Value), nil
}

func (s *Service) markSubTaskRunning(ctx context.Context, taskID string, subTaskSK string, startedAt string) error {
	_, err := s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]dynamotypes.AttributeValue{
			"pk": &dynamotypes.AttributeValueMemberS{Value: taskID},
			"sk": &dynamotypes.AttributeValueMemberS{Value: subTaskSK},
		},
		UpdateExpression:          aws.String("SET #status = :running, started_at = :started_at"),
		ConditionExpression:       aws.String("#status = :queued"),
		ExpressionAttributeNames:  map[string]string{"#status": "status"},
		ExpressionAttributeValues: map[string]dynamotypes.AttributeValue{":queued": &dynamotypes.AttributeValueMemberS{Value: string(models.SubTaskStatusQueued)}, ":running": &dynamotypes.AttributeValueMemberS{Value: string(models.SubTaskStatusRunning)}, ":started_at": &dynamotypes.AttributeValueMemberS{Value: startedAt}},
	})
	if err != nil {
		return fmt.Errorf("mark running: %w", err)
	}
	return nil
}

func (s *Service) publishFailure(ctx context.Context, taskID string, subTaskSK string, err error) error {
	if err == nil {
		return nil
	}
	pubErr := s.publishResult(ctx, models.SubTaskResultMessage{
		TaskID:    taskID,
		SubTaskSK: subTaskSK,
		Error:     err.Error(),
	})
	if pubErr != nil {
		return errors.Join(err, pubErr)
	}
	return nil
}

func (s *Service) publishResult(ctx context.Context, msg models.SubTaskResultMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	if _, err := s.sqs.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.resultsQueue),
		MessageBody: aws.String(string(body)),
	}); err != nil {
		return fmt.Errorf("send result message: %w", err)
	}
	return nil
}

func (s *Service) loadSecureString(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("missing ssm parameter name")
	}

	out, err := s.ssm.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("ssm get parameter %q: %w", name, err)
	}
	if out.Parameter == nil {
		return "", fmt.Errorf("ssm parameter %q missing", name)
	}
	value := strings.TrimSpace(aws.ToString(out.Parameter.Value))
	if value == "" {
		return "", fmt.Errorf("ssm parameter %q empty", name)
	}
	return value, nil
}

func (s *Service) writeRunLog(ctx context.Context, taskID string, subTaskSK string, agentType models.AgentType, eventType string, tokensIn int, tokensOut int, lesserRef string, detail string) error {
	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		return nil
	}
	agentType = models.AgentType(strings.TrimSpace(string(agentType)))
	if agentType == "" {
		agentType = "UNKNOWN"
	}

	now := s.now().UTC()
	ttl := now.Add(30 * 24 * time.Hour).Unix()

	detail = strings.TrimSpace(detail)
	if len(detail) > 2048 {
		detail = detail[:2048]
	}

	item := map[string]dynamotypes.AttributeValue{
		"pk":         &dynamotypes.AttributeValueMemberS{Value: taskID},
		"sk":         &dynamotypes.AttributeValueMemberS{Value: ulid.Make().String()},
		"agent_type": &dynamotypes.AttributeValueMemberS{Value: string(agentType)},
		"event_type": &dynamotypes.AttributeValueMemberS{Value: eventType},
		"ttl":        &dynamotypes.AttributeValueMemberN{Value: strconv.FormatInt(ttl, 10)},
	}

	if strings.TrimSpace(subTaskSK) != "" {
		item["subtask_sk"] = &dynamotypes.AttributeValueMemberS{Value: subTaskSK}
	}
	if tokensIn > 0 {
		item["tokens_in"] = &dynamotypes.AttributeValueMemberN{Value: strconv.Itoa(tokensIn)}
	}
	if tokensOut > 0 {
		item["tokens_out"] = &dynamotypes.AttributeValueMemberN{Value: strconv.Itoa(tokensOut)}
	}
	if strings.TrimSpace(lesserRef) != "" {
		item["lesser_ref"] = &dynamotypes.AttributeValueMemberS{Value: strings.TrimSpace(lesserRef)}
	}
	if detail != "" {
		item["detail"] = &dynamotypes.AttributeValueMemberS{Value: detail}
	}

	if _, err := s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	}); err != nil {
		return fmt.Errorf("put runlog: %w", err)
	}
	return nil
}

type systemPromptData struct {
	Goal      string
	AgentType string
	Memory    []string
}

func normalizeMemoryNotes(mem *lesser.AgentMemorySearchResult) []string {
	if mem == nil || len(mem.Notes) == 0 {
		return nil
	}

	out := make([]string, 0, len(mem.Notes))
	for i, n := range mem.Notes {
		if i >= 10 {
			break
		}
		content := strings.TrimSpace(n.Content)
		if len(content) > 500 {
			content = content[:500]
		}
		if content == "" {
			continue
		}
		out = append(out, content)
	}
	return out
}

func renderSystemPrompt(tpl string, data systemPromptData) (string, error) {
	tpl = strings.TrimSpace(tpl)
	if tpl == "" {
		tpl = "You are an agent.\n\nGoal:\n{{.Goal}}\n"
	}

	t, err := template.New("system_prompt").Option("missingkey=error").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parse system prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute system prompt template: %w", err)
	}
	return buf.String(), nil
}

func isConditionalCheckFailed(err error) bool {
	var ccfe *dynamotypes.ConditionalCheckFailedException
	return errors.As(err, &ccfe)
}

func (s *Service) loadAgentConfig(ctx context.Context, agentType models.AgentType) (*models.AgentConfig, error) {
	pk, err := models.AgentConfigPK(s.instanceDomain)
	if err != nil {
		return nil, err
	}
	sk, err := models.AgentConfigSK(agentType)
	if err != nil {
		return nil, err
	}

	out, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.tableName),
		ConsistentRead: aws.Bool(true),
		Key: map[string]dynamotypes.AttributeValue{
			"pk": &dynamotypes.AttributeValueMemberS{Value: pk},
			"sk": &dynamotypes.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get agent config: %w", err)
	}
	if len(out.Item) == 0 {
		return nil, fmt.Errorf("agent config not found (pk=%s sk=%s)", pk, sk)
	}

	cfg := &models.AgentConfig{
		PK:        pk,
		SK:        sk,
		AgentType: agentType,
	}

	if v, ok := out.Item["enabled"].(*dynamotypes.AttributeValueMemberBOOL); ok {
		cfg.Enabled = v.Value
	}
	if v, ok := out.Item["model_id"].(*dynamotypes.AttributeValueMemberS); ok {
		cfg.ModelID = strings.TrimSpace(v.Value)
	}
	if v, ok := out.Item["max_tokens"].(*dynamotypes.AttributeValueMemberN); ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v.Value)); parseErr == nil {
			cfg.MaxTokens = n
		}
	}
	if v, ok := out.Item["system_prompt_template"].(*dynamotypes.AttributeValueMemberS); ok {
		cfg.SystemPromptTemplate = v.Value
	}
	if v, ok := out.Item["token_ssm_path"].(*dynamotypes.AttributeValueMemberS); ok {
		cfg.TokenSSMPath = strings.TrimSpace(v.Value)
	}
	if v, ok := out.Item["refresh_ssm_path"].(*dynamotypes.AttributeValueMemberS); ok {
		cfg.RefreshSSMPath = strings.TrimSpace(v.Value)
	}
	if v, ok := out.Item["lesser_username"].(*dynamotypes.AttributeValueMemberS); ok {
		cfg.LesserUsername = strings.TrimSpace(v.Value)
	}
	if v, ok := out.Item["queue_url"].(*dynamotypes.AttributeValueMemberS); ok {
		cfg.QueueURL = strings.TrimSpace(v.Value)
	}

	return cfg, nil
}
