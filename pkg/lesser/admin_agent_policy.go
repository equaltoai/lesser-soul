package lesser

import (
	"context"
	"fmt"
)

const adminAgentPolicyQuery = `query AdminAgentPolicy {
  adminAgentPolicy {
    allowAgents
    allowAgentRegistration
    defaultQuarantineDays
    maxAgentsPerOwner
    allowRemoteAgents
    remoteQuarantineDays
    blockedAgentDomains
    trustedAgentDomains
    agentMaxPostsPerHour
    verifiedAgentMaxPostsPerHour
    agentMaxFollowsPerHour
    verifiedAgentMaxFollowsPerHour
    hybridRetrievalEnabled
    hybridRetrievalMaxCandidates
    updatedAt
  }
}`

type AdminAgentPolicy struct {
	AllowAgents                    bool     `json:"allowAgents"`
	AllowAgentRegistration         bool     `json:"allowAgentRegistration"`
	DefaultQuarantineDays          int      `json:"defaultQuarantineDays"`
	MaxAgentsPerOwner              int      `json:"maxAgentsPerOwner"`
	AllowRemoteAgents              bool     `json:"allowRemoteAgents"`
	RemoteQuarantineDays           int      `json:"remoteQuarantineDays"`
	BlockedAgentDomains            []string `json:"blockedAgentDomains"`
	TrustedAgentDomains            []string `json:"trustedAgentDomains"`
	AgentMaxPostsPerHour           int      `json:"agentMaxPostsPerHour"`
	VerifiedAgentMaxPostsPerHour   int      `json:"verifiedAgentMaxPostsPerHour"`
	AgentMaxFollowsPerHour         int      `json:"agentMaxFollowsPerHour"`
	VerifiedAgentMaxFollowsPerHour int      `json:"verifiedAgentMaxFollowsPerHour"`
	HybridRetrievalEnabled         bool     `json:"hybridRetrievalEnabled"`
	HybridRetrievalMaxCandidates   int      `json:"hybridRetrievalMaxCandidates"`
	UpdatedAt                      string   `json:"updatedAt"`
}

func (c *Client) AdminAgentPolicy(ctx context.Context) (*AdminAgentPolicy, error) {
	var resp graphQLResponse[struct {
		AdminAgentPolicy *AdminAgentPolicy `json:"adminAgentPolicy"`
	}]

	if err := c.doGraphQL(ctx, adminAgentPolicyQuery, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", resp.Errors[0].Message)
	}
	if resp.Data.AdminAgentPolicy == nil {
		return nil, fmt.Errorf("graphql: missing adminAgentPolicy response")
	}
	return resp.Data.AdminAgentPolicy, nil
}

const updateAdminAgentPolicyMutation = `mutation UpdateAdminAgentPolicy($input: UpdateAdminAgentPolicyInput!) {
  updateAdminAgentPolicy(input: $input) {
    allowAgents
    allowAgentRegistration
    defaultQuarantineDays
    maxAgentsPerOwner
    allowRemoteAgents
    remoteQuarantineDays
    blockedAgentDomains
    trustedAgentDomains
    agentMaxPostsPerHour
    verifiedAgentMaxPostsPerHour
    agentMaxFollowsPerHour
    verifiedAgentMaxFollowsPerHour
    hybridRetrievalEnabled
    hybridRetrievalMaxCandidates
    updatedAt
  }
}`

type UpdateAdminAgentPolicyInput struct {
	AllowAgents                    bool     `json:"allowAgents"`
	AllowAgentRegistration         bool     `json:"allowAgentRegistration"`
	DefaultQuarantineDays          int      `json:"defaultQuarantineDays"`
	MaxAgentsPerOwner              int      `json:"maxAgentsPerOwner"`
	AllowRemoteAgents              bool     `json:"allowRemoteAgents"`
	RemoteQuarantineDays           int      `json:"remoteQuarantineDays"`
	BlockedAgentDomains            []string `json:"blockedAgentDomains,omitempty"`
	TrustedAgentDomains            []string `json:"trustedAgentDomains,omitempty"`
	AgentMaxPostsPerHour           int      `json:"agentMaxPostsPerHour"`
	VerifiedAgentMaxPostsPerHour   int      `json:"verifiedAgentMaxPostsPerHour"`
	AgentMaxFollowsPerHour         int      `json:"agentMaxFollowsPerHour"`
	VerifiedAgentMaxFollowsPerHour int      `json:"verifiedAgentMaxFollowsPerHour"`
	HybridRetrievalEnabled         bool     `json:"hybridRetrievalEnabled"`
	HybridRetrievalMaxCandidates   int      `json:"hybridRetrievalMaxCandidates"`
}

func (c *Client) UpdateAdminAgentPolicy(ctx context.Context, input UpdateAdminAgentPolicyInput) (*AdminAgentPolicy, error) {
	vars := struct {
		Input UpdateAdminAgentPolicyInput `json:"input"`
	}{Input: input}

	var resp graphQLResponse[struct {
		UpdateAdminAgentPolicy *AdminAgentPolicy `json:"updateAdminAgentPolicy"`
	}]

	if err := c.doGraphQL(ctx, updateAdminAgentPolicyMutation, vars, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", resp.Errors[0].Message)
	}
	if resp.Data.UpdateAdminAgentPolicy == nil {
		return nil, fmt.Errorf("graphql: missing updateAdminAgentPolicy response")
	}
	return resp.Data.UpdateAdminAgentPolicy, nil
}
