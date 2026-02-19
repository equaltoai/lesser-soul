package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/equaltoai/lesser-soul/pkg/config"
	"github.com/equaltoai/lesser-soul/pkg/lesser"
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
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  %s bootstrap-researcher [flags]\n", os.Args[0])
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
