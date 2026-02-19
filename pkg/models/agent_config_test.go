package models

import "testing"

func TestAgentConfigPKAndSK(t *testing.T) {
	t.Parallel()

	pk, err := AgentConfigPK("dev.example.com")
	if err != nil {
		t.Fatalf("AgentConfigPK() err=%v", err)
	}
	if pk != "CONFIG#dev.example.com" {
		t.Fatalf("pk=%q want=%q", pk, "CONFIG#dev.example.com")
	}

	sk, err := AgentConfigSK(AgentTypeResearcher)
	if err != nil {
		t.Fatalf("AgentConfigSK() err=%v", err)
	}
	if sk != "AGENT#RESEARCHER" {
		t.Fatalf("sk=%q want=%q", sk, "AGENT#RESEARCHER")
	}
}

func TestAgentConfigPK_Empty(t *testing.T) {
	t.Parallel()

	if _, err := AgentConfigPK(""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAgentConfigSK_Empty(t *testing.T) {
	t.Parallel()

	if _, err := AgentConfigSK(""); err == nil {
		t.Fatalf("expected error")
	}
}
