package models

import (
	"fmt"
	"strings"
)

const agentConfigPKPrefix = "CONFIG#"
const agentConfigSKPrefix = "AGENT#"

func AgentConfigPK(instanceDomain string) (string, error) {
	instanceDomain = strings.TrimSpace(instanceDomain)
	if instanceDomain == "" {
		return "", fmt.Errorf("missing instanceDomain")
	}
	return agentConfigPKPrefix + instanceDomain, nil
}

func AgentConfigSK(agentType AgentType) (string, error) {
	agentType = AgentType(strings.TrimSpace(string(agentType)))
	if agentType == "" {
		return "", fmt.Errorf("missing agentType")
	}
	return agentConfigSKPrefix + string(agentType), nil
}

type AgentConfig struct {
	PK                   string    `json:"pk"`
	SK                   string    `json:"sk"`
	AgentType            AgentType `json:"agent_type"`
	Enabled              bool      `json:"enabled"`
	ModelID              string    `json:"model_id"`
	MaxTokens            int       `json:"max_tokens"`
	SystemPromptTemplate string    `json:"system_prompt_template"`
	TokenSSMPath         string    `json:"token_ssm_path"`
	RefreshSSMPath       string    `json:"refresh_ssm_path,omitempty"`
	LesserUsername       string    `json:"lesser_username"`
	QueueURL             string    `json:"queue_url"`
}
