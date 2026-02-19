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
