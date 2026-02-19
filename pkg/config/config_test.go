package config

import "testing"

func TestParseStage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    Stage
		wantErr bool
	}{
		{name: "lab", in: "lab", want: StageLab},
		{name: "live", in: "live", want: StageLive},
		{name: "trims", in: " live ", want: StageLive},
		{name: "case-insensitive", in: "LAB", want: StageLab},
		{name: "invalid", in: "prod", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseStage(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseStage(%q) err=%v wantErr=%v", tt.in, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Fatalf("ParseStage(%q)=%q want=%q", tt.in, got, tt.want)
			}
		})
	}
}

func TestStageFromEnv(t *testing.T) {
	t.Setenv(EnvSoulStage, "lab")
	got, err := StageFromEnv()
	if err != nil {
		t.Fatalf("StageFromEnv() err=%v", err)
	}
	if got != StageLab {
		t.Fatalf("StageFromEnv()=%q want=%q", got, StageLab)
	}
}

func TestSSMPaths(t *testing.T) {
	t.Parallel()

	const domain = "dev.simulacrum.greater.website"

	base, err := SSMBasePath(domain)
	if err != nil {
		t.Fatalf("SSMBasePath() err=%v", err)
	}
	if base != "/soul/dev.simulacrum.greater.website" {
		t.Fatalf("SSMBasePath()=%q", base)
	}

	p, err := SSMPath(domain, "inference", "url")
	if err != nil {
		t.Fatalf("SSMPath() err=%v", err)
	}
	if p != "/soul/dev.simulacrum.greater.website/inference/url" {
		t.Fatalf("SSMPath()=%q", p)
	}

	p2, err := InferenceKeySSMPath(domain)
	if err != nil {
		t.Fatalf("InferenceKeySSMPath() err=%v", err)
	}
	if p2 != "/soul/dev.simulacrum.greater.website/inference/key" {
		t.Fatalf("InferenceKeySSMPath()=%q", p2)
	}

	p3, err := LesserHostInstanceKeySSMPath(domain)
	if err != nil {
		t.Fatalf("LesserHostInstanceKeySSMPath() err=%v", err)
	}
	if p3 != "/soul/dev.simulacrum.greater.website/lesser-host/instance-key" {
		t.Fatalf("LesserHostInstanceKeySSMPath()=%q", p3)
	}
}

func TestSSMBasePathRejectsScheme(t *testing.T) {
	t.Parallel()

	if _, err := SSMBasePath("https://simulacrum.greater.website"); err == nil {
		t.Fatalf("SSMBasePath() expected error")
	}
}

func TestLesserGraphQLURLFromEnv(t *testing.T) {
	t.Setenv(EnvLesserGraphQLURL, "https://example.com/api/graphql")
	got, err := LesserGraphQLURLFromEnv()
	if err != nil {
		t.Fatalf("LesserGraphQLURLFromEnv() err=%v", err)
	}
	if got != "https://example.com/api/graphql" {
		t.Fatalf("LesserGraphQLURLFromEnv()=%q", got)
	}
}

func TestStateTableNameFromEnv(t *testing.T) {
	t.Setenv(EnvSoulStateTableName, "soul-lab")
	got, err := StateTableNameFromEnv()
	if err != nil {
		t.Fatalf("StateTableNameFromEnv() err=%v", err)
	}
	if got != "soul-lab" {
		t.Fatalf("StateTableNameFromEnv()=%q", got)
	}
}

func TestQueueURLsFromEnv(t *testing.T) {
	t.Setenv(EnvSoulResearcherQueueURL, "https://sqs.example.com/researcher")
	got, err := ResearcherQueueURLFromEnv()
	if err != nil {
		t.Fatalf("ResearcherQueueURLFromEnv() err=%v", err)
	}
	if got == "" {
		t.Fatalf("ResearcherQueueURLFromEnv() empty")
	}

	t.Setenv(EnvSoulAssistantQueueURL, "https://sqs.example.com/assistant")
	gotAssistant, err := AssistantQueueURLFromEnv()
	if err != nil {
		t.Fatalf("AssistantQueueURLFromEnv() err=%v", err)
	}
	if gotAssistant == "" {
		t.Fatalf("AssistantQueueURLFromEnv() empty")
	}

	t.Setenv(EnvSoulCuratorQueueURL, "https://sqs.example.com/curator")
	gotCurator, err := CuratorQueueURLFromEnv()
	if err != nil {
		t.Fatalf("CuratorQueueURLFromEnv() err=%v", err)
	}
	if gotCurator == "" {
		t.Fatalf("CuratorQueueURLFromEnv() empty")
	}

	t.Setenv(EnvSoulCustomCoderQueueURL, "https://sqs.example.com/custom-coder")
	gotCustomCoder, err := CustomCoderQueueURLFromEnv()
	if err != nil {
		t.Fatalf("CustomCoderQueueURLFromEnv() err=%v", err)
	}
	if gotCustomCoder == "" {
		t.Fatalf("CustomCoderQueueURLFromEnv() empty")
	}

	t.Setenv(EnvSoulCustomSummarizerQueueURL, "https://sqs.example.com/custom-summarizer")
	gotCustomSummarizer, err := CustomSummarizerQueueURLFromEnv()
	if err != nil {
		t.Fatalf("CustomSummarizerQueueURLFromEnv() err=%v", err)
	}
	if gotCustomSummarizer == "" {
		t.Fatalf("CustomSummarizerQueueURLFromEnv() empty")
	}

	t.Setenv(EnvSoulResultsQueueURL, "https://sqs.example.com/results")
	got2, err := ResultsQueueURLFromEnv()
	if err != nil {
		t.Fatalf("ResultsQueueURLFromEnv() err=%v", err)
	}
	if got2 == "" {
		t.Fatalf("ResultsQueueURLFromEnv() empty")
	}
}

func TestAgentTokenSSMPaths(t *testing.T) {
	t.Parallel()

	const domain = "dev.simulacrum.greater.website"

	p, err := AgentTokenSSMPath(domain, "researcher")
	if err != nil {
		t.Fatalf("AgentTokenSSMPath() err=%v", err)
	}
	if p != "/soul/dev.simulacrum.greater.website/agents/researcher/token" {
		t.Fatalf("AgentTokenSSMPath()=%q", p)
	}

	p2, err := AgentRefreshSSMPath(domain, "researcher")
	if err != nil {
		t.Fatalf("AgentRefreshSSMPath() err=%v", err)
	}
	if p2 != "/soul/dev.simulacrum.greater.website/agents/researcher/refresh" {
		t.Fatalf("AgentRefreshSSMPath()=%q", p2)
	}
}
