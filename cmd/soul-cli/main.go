package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/equaltoai/lesser-soul/pkg/config"
	"github.com/equaltoai/lesser-soul/pkg/lesser"
	"github.com/equaltoai/lesser-soul/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "bootstrap-researcher":
		if err := bootstrapResearcher(os.Args[2:]); err != nil {
			log.Fatalf("bootstrap-researcher: %v", err)
		}
	case "bootstrap-agents":
		if err := bootstrapAgents(os.Args[2:]); err != nil {
			log.Fatalf("bootstrap-agents: %v", err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  %s bootstrap-researcher [flags]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s bootstrap-agents [flags]\n", os.Args[0])
}

func bootstrapResearcher(args []string) error {
	fs := flag.NewFlagSet("bootstrap-researcher", flag.ContinueOnError)

	instanceDomain := fs.String("instance-domain", strings.TrimSpace(os.Getenv(config.EnvSoulInstanceDomain)), "Instance domain (e.g. dev.simulacrum.greater.website)")
	graphQLURL := fs.String("graphql-url", strings.TrimSpace(os.Getenv(config.EnvLesserGraphQLURL)), "Lesser GraphQL URL (e.g. https://<domain>/api/graphql)")
	jwtSecretID := fs.String("jwt-secret-id", "", "AWS Secrets Manager secret id for Lesser JWT secret (optional if --admin-bearer-token is set)")
	adminBearerToken := fs.String("admin-bearer-token", "", "Operator bearer token for Lesser GraphQL (optional if --jwt-secret-id is set)")
	operatorUsername := fs.String("operator-username", "soul-bootstrap", "Username for minted operator JWT (only used with --jwt-secret-id)")

	agentUsername := fs.String("agent-username", "soul-researcher", "Agent username to create/delegate")
	agentDisplayName := fs.String("agent-display-name", "Soul Researcher", "Agent display name")
	agentBio := fs.String("agent-bio", "Research agent", "Agent bio")
	agentVersion := fs.String("agent-version", "lesser-soul/0.1.0", "Agent version string")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*instanceDomain) == "" {
		return fmt.Errorf("missing --instance-domain (or %s)", config.EnvSoulInstanceDomain)
	}
	if strings.TrimSpace(*graphQLURL) == "" {
		return fmt.Errorf("missing --graphql-url (or %s)", config.EnvLesserGraphQLURL)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bearer := strings.TrimSpace(*adminBearerToken)
	if bearer == "" {
		secretID := strings.TrimSpace(*jwtSecretID)
		if secretID == "" {
			return fmt.Errorf("missing auth: set --admin-bearer-token or --jwt-secret-id")
		}

		awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("load aws config: %w", err)
		}
		sm := secretsmanager.NewFromConfig(awsCfg)

		jwtSecret, err := loadJWTSecret(ctx, sm, secretID)
		if err != nil {
			return err
		}

		bearer, err = mintOperatorJWT(jwtSecret, strings.TrimSpace(*operatorUsername), []string{"read", "write", "admin"}, "soul-bootstrap", "cli")
		if err != nil {
			return err
		}
	}

	lc, err := lesser.NewClient(*graphQLURL, lesser.WithBearerToken(bearer))
	if err != nil {
		return err
	}

	policy, err := lc.AdminAgentPolicy(ctx)
	if err != nil {
		return err
	}
	if !policy.AllowAgents || !policy.AllowAgentRegistration {
		_, err := lc.UpdateAdminAgentPolicy(ctx, lesser.UpdateAdminAgentPolicyInput{
			AllowAgents:                    true,
			AllowAgentRegistration:         true,
			DefaultQuarantineDays:          policy.DefaultQuarantineDays,
			MaxAgentsPerOwner:              policy.MaxAgentsPerOwner,
			AllowRemoteAgents:              policy.AllowRemoteAgents,
			RemoteQuarantineDays:           policy.RemoteQuarantineDays,
			BlockedAgentDomains:            policy.BlockedAgentDomains,
			TrustedAgentDomains:            policy.TrustedAgentDomains,
			AgentMaxPostsPerHour:           policy.AgentMaxPostsPerHour,
			VerifiedAgentMaxPostsPerHour:   policy.VerifiedAgentMaxPostsPerHour,
			AgentMaxFollowsPerHour:         policy.AgentMaxFollowsPerHour,
			VerifiedAgentMaxFollowsPerHour: policy.VerifiedAgentMaxFollowsPerHour,
			HybridRetrievalEnabled:         policy.HybridRetrievalEnabled,
			HybridRetrievalMaxCandidates:   policy.HybridRetrievalMaxCandidates,
		})
		if err != nil {
			return err
		}
		log.Printf("enabled agent support via admin policy (allowAgents=true allowAgentRegistration=true)")
	}

	delegation, err := lc.DelegateToAgent(ctx, lesser.DelegateToAgentInput{
		AgentUsername: strings.TrimSpace(*agentUsername),
		DisplayName:   strings.TrimSpace(*agentDisplayName),
		Bio:           strings.TrimSpace(*agentBio),
		Scopes:        []string{"read", "write"},
		AgentType:     "RESEARCHER",
		Version:       strings.TrimSpace(*agentVersion),
	})
	if err != nil {
		return err
	}

	if _, err := lc.AdminVerifyAgent(ctx, strings.TrimSpace(*agentUsername), &lesser.AdminVerifyAgentInput{
		Reason:         "soul bootstrap",
		ExitQuarantine: true,
	}); err != nil {
		return err
	}

	agent, err := lc.Agent(ctx, strings.TrimSpace(*agentUsername))
	if err != nil {
		return err
	}

	tokenPath, err := config.AgentTokenSSMPath(*instanceDomain, "researcher")
	if err != nil {
		return err
	}
	refreshPath, err := config.AgentRefreshSSMPath(*instanceDomain, "researcher")
	if err != nil {
		return err
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}
	ssmClient := ssm.NewFromConfig(awsCfg)

	if err := putSecureString(ctx, ssmClient, tokenPath, delegation.AccessToken); err != nil {
		return err
	}
	if err := putSecureString(ctx, ssmClient, refreshPath, delegation.RefreshToken); err != nil {
		return err
	}

	log.Printf("wrote researcher tokens to ssm (token=%s refresh=%s expiresIn=%ds verified=%v)", tokenPath, refreshPath, delegation.ExpiresIn, agent.Verified)
	return nil
}

type agentBootstrapSpec struct {
	AgentType            models.AgentType
	AgentSlug            string
	AgentUsername        string
	DisplayName          string
	Bio                  string
	ModelID              string
	MaxTokens            int
	SystemPromptTemplate string
}

func bootstrapAgents(args []string) error {
	fs := flag.NewFlagSet("bootstrap-agents", flag.ContinueOnError)

	instanceDomain := fs.String("instance-domain", strings.TrimSpace(os.Getenv(config.EnvSoulInstanceDomain)), "Instance domain (e.g. dev.simulacrum.greater.website)")
	graphQLURL := fs.String("graphql-url", strings.TrimSpace(os.Getenv(config.EnvLesserGraphQLURL)), "Lesser GraphQL URL (e.g. https://<domain>/api/graphql)")
	stageRaw := fs.String("stage", strings.TrimSpace(os.Getenv(config.EnvSoulStage)), "Soul stage (lab|live)")
	tableName := fs.String("table-name", "", "DynamoDB table name (default: soul-<stage>)")

	jwtSecretID := fs.String("jwt-secret-id", "", "AWS Secrets Manager secret id for Lesser JWT secret (optional if --admin-bearer-token is set)")
	adminBearerToken := fs.String("admin-bearer-token", "", "Operator bearer token for Lesser GraphQL (optional if --jwt-secret-id is set)")
	operatorUsername := fs.String("operator-username", "soul-bootstrap", "Username for minted operator JWT (only used with --jwt-secret-id)")

	agentVersion := fs.String("agent-version", "lesser-soul/0.2.0", "Agent version string")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*instanceDomain) == "" {
		return fmt.Errorf("missing --instance-domain (or %s)", config.EnvSoulInstanceDomain)
	}
	if strings.TrimSpace(*graphQLURL) == "" {
		return fmt.Errorf("missing --graphql-url (or %s)", config.EnvLesserGraphQLURL)
	}
	if strings.TrimSpace(*stageRaw) == "" {
		return fmt.Errorf("missing --stage (or %s)", config.EnvSoulStage)
	}
	stage, err := config.ParseStage(*stageRaw)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*tableName) == "" {
		*tableName = fmt.Sprintf("soul-%s", stage)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}
	sm := secretsmanager.NewFromConfig(awsCfg)
	ssmClient := ssm.NewFromConfig(awsCfg)
	sqsClient := sqs.NewFromConfig(awsCfg)
	db := dynamodb.NewFromConfig(awsCfg)

	bearer := strings.TrimSpace(*adminBearerToken)
	if bearer == "" {
		secretID := strings.TrimSpace(*jwtSecretID)
		if secretID == "" {
			return fmt.Errorf("missing auth: set --admin-bearer-token or --jwt-secret-id")
		}

		jwtSecret, err := loadJWTSecret(ctx, sm, secretID)
		if err != nil {
			return err
		}

		bearer, err = mintOperatorJWT(jwtSecret, strings.TrimSpace(*operatorUsername), []string{"read", "write", "admin"}, "soul-bootstrap", "cli")
		if err != nil {
			return err
		}
	}

	lc, err := lesser.NewClient(*graphQLURL, lesser.WithBearerToken(bearer))
	if err != nil {
		return err
	}

	policy, err := lc.AdminAgentPolicy(ctx)
	if err != nil {
		return err
	}
	if !policy.AllowAgents || !policy.AllowAgentRegistration {
		_, err := lc.UpdateAdminAgentPolicy(ctx, lesser.UpdateAdminAgentPolicyInput{
			AllowAgents:                    true,
			AllowAgentRegistration:         true,
			DefaultQuarantineDays:          policy.DefaultQuarantineDays,
			MaxAgentsPerOwner:              policy.MaxAgentsPerOwner,
			AllowRemoteAgents:              policy.AllowRemoteAgents,
			RemoteQuarantineDays:           policy.RemoteQuarantineDays,
			BlockedAgentDomains:            policy.BlockedAgentDomains,
			TrustedAgentDomains:            policy.TrustedAgentDomains,
			AgentMaxPostsPerHour:           policy.AgentMaxPostsPerHour,
			VerifiedAgentMaxPostsPerHour:   policy.VerifiedAgentMaxPostsPerHour,
			AgentMaxFollowsPerHour:         policy.AgentMaxFollowsPerHour,
			VerifiedAgentMaxFollowsPerHour: policy.VerifiedAgentMaxFollowsPerHour,
			HybridRetrievalEnabled:         policy.HybridRetrievalEnabled,
			HybridRetrievalMaxCandidates:   policy.HybridRetrievalMaxCandidates,
		})
		if err != nil {
			return err
		}
		log.Printf("enabled agent support via admin policy (allowAgents=true allowAgentRegistration=true)")
	}

	agents := []agentBootstrapSpec{
		{
			AgentType:            models.AgentTypeResearcher,
			AgentSlug:            "researcher",
			AgentUsername:        "soul-researcher",
			DisplayName:          "Soul Researcher",
			Bio:                  "Research agent",
			ModelID:              "llama-3.3-70b",
			MaxTokens:            1200,
			SystemPromptTemplate: "You are a research agent.\nUse the memory notes below if relevant. If irrelevant, ignore them.\n\nGoal:\n{{.Goal}}\n\nMemory:\n{{range .Memory}}- {{.}}\n{{end}}",
		},
		{
			AgentType:            models.AgentTypeAssistant,
			AgentSlug:            "assistant",
			AgentUsername:        "soul-assistant",
			DisplayName:          "Soul Assistant",
			Bio:                  "General assistant agent",
			ModelID:              "qwen2.5-72b-instruct",
			MaxTokens:            900,
			SystemPromptTemplate: "You are a helpful assistant.\n\nGoal:\n{{.Goal}}\n\nMemory:\n{{range .Memory}}- {{.}}\n{{end}}",
		},
		{
			AgentType:            models.AgentTypeCurator,
			AgentSlug:            "curator",
			AgentUsername:        "soul-curator",
			DisplayName:          "Soul Curator",
			Bio:                  "Summarization and curation agent",
			ModelID:              "llama-3.1-8b",
			MaxTokens:            600,
			SystemPromptTemplate: "You are a curator agent. Produce concise, high-signal output.\n\nGoal:\n{{.Goal}}\n\nMemory:\n{{range .Memory}}- {{.}}\n{{end}}",
		},
		{
			AgentType:            models.AgentTypeCustomCoder,
			AgentSlug:            "custom-coder",
			AgentUsername:        "soul-custom-coder",
			DisplayName:          "Soul Custom Coder",
			Bio:                  "Custom coding agent",
			ModelID:              "qwen2.5-coder-32b",
			MaxTokens:            1200,
			SystemPromptTemplate: "You are a coding agent. Provide correct, runnable code when asked.\n\nGoal:\n{{.Goal}}\n\nMemory:\n{{range .Memory}}- {{.}}\n{{end}}",
		},
		{
			AgentType:            models.AgentTypeCustomSummarizer,
			AgentSlug:            "custom-summarizer",
			AgentUsername:        "soul-custom-summarizer",
			DisplayName:          "Soul Custom Summarizer",
			Bio:                  "Custom summarization agent",
			ModelID:              "phi-3.5-mini",
			MaxTokens:            700,
			SystemPromptTemplate: "You are a summarizer agent. Produce a tight summary.\n\nGoal:\n{{.Goal}}\n\nMemory:\n{{range .Memory}}- {{.}}\n{{end}}",
		},
	}

	for _, a := range agents {
		if err := bootstrapSingleAgent(ctx, lc, ssmClient, sqsClient, db, *tableName, *instanceDomain, stage, strings.TrimSpace(*agentVersion), a); err != nil {
			return err
		}
	}

	return nil
}

func bootstrapSingleAgent(ctx context.Context, lc *lesser.Client, ssmClient *ssm.Client, sqsClient *sqs.Client, db *dynamodb.Client, tableName string, instanceDomain string, stage config.Stage, version string, spec agentBootstrapSpec) error {
	queueURL, err := resolveAgentQueueURL(ctx, sqsClient, stage, spec.AgentType)
	if err != nil {
		return err
	}

	delegation, err := lc.DelegateToAgent(ctx, lesser.DelegateToAgentInput{
		AgentUsername: strings.TrimSpace(spec.AgentUsername),
		DisplayName:   strings.TrimSpace(spec.DisplayName),
		Bio:           strings.TrimSpace(spec.Bio),
		Scopes:        []string{"read", "write"},
		AgentType:     lesserAgentType(spec.AgentType),
		Version:       version,
	})
	if err != nil {
		return err
	}

	if _, err := lc.AdminVerifyAgent(ctx, strings.TrimSpace(spec.AgentUsername), &lesser.AdminVerifyAgentInput{
		Reason:         "soul bootstrap",
		ExitQuarantine: true,
	}); err != nil {
		return err
	}

	agent, err := lc.Agent(ctx, strings.TrimSpace(spec.AgentUsername))
	if err != nil {
		return err
	}

	tokenPath, err := config.AgentTokenSSMPath(instanceDomain, spec.AgentSlug)
	if err != nil {
		return err
	}
	refreshPath, err := config.AgentRefreshSSMPath(instanceDomain, spec.AgentSlug)
	if err != nil {
		return err
	}

	if err := putSecureString(ctx, ssmClient, tokenPath, delegation.AccessToken); err != nil {
		return err
	}
	if err := putSecureString(ctx, ssmClient, refreshPath, delegation.RefreshToken); err != nil {
		return err
	}

	if err := upsertAgentConfig(ctx, db, tableName, instanceDomain, spec.AgentType, queueURL, tokenPath, refreshPath, spec); err != nil {
		return err
	}

	log.Printf("bootstrapped agent=%s (type=%s verified=%v expiresIn=%ds token=%s refresh=%s queue=%s table=%s)", spec.AgentUsername, spec.AgentType, agent.Verified, delegation.ExpiresIn, tokenPath, refreshPath, queueURL, tableName)
	return nil
}

func lesserAgentType(agentType models.AgentType) string {
	switch agentType {
	case models.AgentTypeCustomCoder, models.AgentTypeCustomSummarizer:
		return "CUSTOM"
	default:
		return string(agentType)
	}
}

func resolveAgentQueueURL(ctx context.Context, api *sqs.Client, stage config.Stage, agentType models.AgentType) (string, error) {
	queueName := ""
	switch agentType {
	case models.AgentTypeResearcher:
		queueName = fmt.Sprintf("soul-researcher-%s", stage)
	case models.AgentTypeAssistant:
		queueName = fmt.Sprintf("soul-assistant-%s", stage)
	case models.AgentTypeCurator:
		queueName = fmt.Sprintf("soul-curator-%s", stage)
	case models.AgentTypeCustomCoder:
		queueName = fmt.Sprintf("soul-custom-coder-%s", stage)
	case models.AgentTypeCustomSummarizer:
		queueName = fmt.Sprintf("soul-custom-summarizer-%s", stage)
	default:
		return "", fmt.Errorf("unsupported agentType %q", agentType)
	}

	out, err := api.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String(queueName)})
	if err != nil {
		return "", fmt.Errorf("sqs get-queue-url %q: %w", queueName, err)
	}
	queueURL := strings.TrimSpace(awsString(out.QueueUrl))
	if queueURL == "" {
		return "", fmt.Errorf("sqs queue url for %q was empty", queueName)
	}
	return queueURL, nil
}

func upsertAgentConfig(ctx context.Context, api *dynamodb.Client, tableName string, instanceDomain string, agentType models.AgentType, queueURL string, tokenPath string, refreshPath string, spec agentBootstrapSpec) error {
	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		return fmt.Errorf("missing tableName")
	}

	pk, err := models.AgentConfigPK(instanceDomain)
	if err != nil {
		return err
	}
	sk, err := models.AgentConfigSK(agentType)
	if err != nil {
		return err
	}

	item := map[string]dynamotypes.AttributeValue{
		"pk":                     &dynamotypes.AttributeValueMemberS{Value: pk},
		"sk":                     &dynamotypes.AttributeValueMemberS{Value: sk},
		"agent_type":             &dynamotypes.AttributeValueMemberS{Value: string(agentType)},
		"enabled":                &dynamotypes.AttributeValueMemberBOOL{Value: true},
		"model_id":               &dynamotypes.AttributeValueMemberS{Value: strings.TrimSpace(spec.ModelID)},
		"max_tokens":             &dynamotypes.AttributeValueMemberN{Value: strconv.Itoa(spec.MaxTokens)},
		"system_prompt_template": &dynamotypes.AttributeValueMemberS{Value: strings.TrimSpace(spec.SystemPromptTemplate)},
		"token_ssm_path":         &dynamotypes.AttributeValueMemberS{Value: strings.TrimSpace(tokenPath)},
		"refresh_ssm_path":       &dynamotypes.AttributeValueMemberS{Value: strings.TrimSpace(refreshPath)},
		"lesser_username":        &dynamotypes.AttributeValueMemberS{Value: strings.TrimSpace(spec.AgentUsername)},
		"queue_url":              &dynamotypes.AttributeValueMemberS{Value: strings.TrimSpace(queueURL)},
	}

	if _, err := api.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	}); err != nil {
		return fmt.Errorf("dynamodb put agent config: %w", err)
	}
	return nil
}

func putSecureString(ctx context.Context, api *ssm.Client, name string, value string) error {
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "" {
		return fmt.Errorf("missing ssm parameter name")
	}
	if value == "" {
		return fmt.Errorf("ssm parameter %q value empty", name)
	}

	_, err := api.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(name),
		Type:      ssmtypes.ParameterTypeSecureString,
		Value:     aws.String(value),
		Overwrite: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("ssm put parameter %q: %w", name, err)
	}
	return nil
}

func loadJWTSecret(ctx context.Context, api *secretsmanager.Client, secretID string) (string, error) {
	out, err := api.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &secretID})
	if err != nil {
		return "", fmt.Errorf("secretsmanager get-secret-value %q: %w", secretID, err)
	}
	raw := strings.TrimSpace(awsString(out.SecretString))
	if raw == "" {
		return "", fmt.Errorf("secret %q was empty", secretID)
	}

	if strings.HasPrefix(raw, "{") {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
			if v, ok := decoded["secret"].(string); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v), nil
			}
		}
	}

	return raw, nil
}

func awsString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type operatorClaims struct {
	jwt.RegisteredClaims
	Username    string   `json:"username"`
	Scopes      []string `json:"scopes"`
	ClientID    string   `json:"client_id"`
	ClientClass string   `json:"client_class,omitempty"`
}

func mintOperatorJWT(jwtSecret string, username string, scopes []string, clientID string, clientClass string) (string, error) {
	jwtSecret = strings.TrimSpace(jwtSecret)
	if jwtSecret == "" {
		return "", fmt.Errorf("missing jwtSecret")
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return "", fmt.Errorf("missing username")
	}
	if strings.TrimSpace(clientID) == "" {
		clientID = "soul-bootstrap"
	}

	now := time.Now().UTC()
	claims := operatorClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
			ID:        ulid.Make().String(),
		},
		Username:    username,
		Scopes:      append([]string(nil), scopes...),
		ClientID:    clientID,
		ClientClass: strings.TrimSpace(clientClass),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}
