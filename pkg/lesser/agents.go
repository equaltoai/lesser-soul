package lesser

import (
	"context"
	"fmt"
	"strings"
)

const delegateToAgentMutation = `mutation DelegateToAgent($input: DelegateToAgentInput!) {
  delegateToAgent(input: $input) {
    accessToken
    refreshToken
    expiresIn
    agent {
      username
      verified
    }
  }
}`

type DelegateToAgentInput struct {
	AgentUsername string   `json:"agentUsername"`
	DisplayName   string   `json:"displayName"`
	Bio           string   `json:"bio,omitempty"`
	Scopes        []string `json:"scopes"`
	ExpiresIn     int      `json:"expiresIn,omitempty"`
	AgentType     string   `json:"agentType"`
	AgentVersion  string   `json:"agentVersion,omitempty"`
	Version       string   `json:"version"`
}

type Agent struct {
	Username string `json:"username"`
	Verified bool   `json:"verified"`
}

type DelegationPayload struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	Agent        *Agent `json:"agent"`
}

func (c *Client) DelegateToAgent(ctx context.Context, input DelegateToAgentInput) (*DelegationPayload, error) {
	input.AgentUsername = strings.TrimSpace(input.AgentUsername)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Bio = strings.TrimSpace(input.Bio)
	input.AgentType = strings.TrimSpace(input.AgentType)
	input.AgentVersion = strings.TrimSpace(input.AgentVersion)
	input.Version = strings.TrimSpace(input.Version)

	vars := struct {
		Input DelegateToAgentInput `json:"input"`
	}{Input: input}

	var resp graphQLResponse[struct {
		DelegateToAgent *DelegationPayload `json:"delegateToAgent"`
	}]

	if err := c.doGraphQL(ctx, delegateToAgentMutation, vars, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", resp.Errors[0].Message)
	}
	if resp.Data.DelegateToAgent == nil {
		return nil, fmt.Errorf("graphql: missing delegateToAgent response")
	}
	if strings.TrimSpace(resp.Data.DelegateToAgent.AccessToken) == "" || strings.TrimSpace(resp.Data.DelegateToAgent.RefreshToken) == "" {
		return nil, fmt.Errorf("graphql: missing delegation tokens")
	}
	return resp.Data.DelegateToAgent, nil
}

const adminVerifyAgentMutation = `mutation AdminVerifyAgent($username: String!, $input: AdminVerifyAgentInput) {
  adminVerifyAgent(username: $username, input: $input) {
    username
    verified
  }
}`

type AdminVerifyAgentInput struct {
	Reason         string `json:"reason,omitempty"`
	ExitQuarantine bool   `json:"exitQuarantine,omitempty"`
}

func (c *Client) AdminVerifyAgent(ctx context.Context, username string, input *AdminVerifyAgentInput) (*Agent, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("missing username")
	}
	vars := struct {
		Username string                 `json:"username"`
		Input    *AdminVerifyAgentInput `json:"input,omitempty"`
	}{
		Username: username,
		Input:    input,
	}

	var resp graphQLResponse[struct {
		AdminVerifyAgent *Agent `json:"adminVerifyAgent"`
	}]

	if err := c.doGraphQL(ctx, adminVerifyAgentMutation, vars, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", resp.Errors[0].Message)
	}
	if resp.Data.AdminVerifyAgent == nil {
		return nil, fmt.Errorf("graphql: missing adminVerifyAgent response")
	}
	return resp.Data.AdminVerifyAgent, nil
}

const agentQuery = `query Agent($username: String!) {
  agent(username: $username) {
    username
    verified
  }
}`

func (c *Client) Agent(ctx context.Context, username string) (*Agent, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("missing username")
	}
	vars := struct {
		Username string `json:"username"`
	}{Username: username}

	var resp graphQLResponse[struct {
		Agent *Agent `json:"agent"`
	}]

	if err := c.doGraphQL(ctx, agentQuery, vars, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", resp.Errors[0].Message)
	}
	if resp.Data.Agent == nil {
		return nil, fmt.Errorf("graphql: missing agent response")
	}
	return resp.Data.Agent, nil
}
