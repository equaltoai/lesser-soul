package config

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	EnvSoulStage          = "SOUL_STAGE"
	EnvSoulInstanceDomain = "SOUL_INSTANCE_DOMAIN"
)

type Stage string

const (
	StageLab  Stage = "lab"
	StageLive Stage = "live"
)

func ParseStage(raw string) (Stage, error) {
	stage := Stage(strings.ToLower(strings.TrimSpace(raw)))
	switch stage {
	case StageLab, StageLive:
		return stage, nil
	default:
		return "", fmt.Errorf("invalid %s=%q (expected %q or %q)", EnvSoulStage, raw, StageLab, StageLive)
	}
}

func StageFromEnv() (Stage, error) {
	raw := strings.TrimSpace(os.Getenv(EnvSoulStage))
	if raw == "" {
		return "", fmt.Errorf("missing %s (expected %q or %q)", EnvSoulStage, StageLab, StageLive)
	}
	return ParseStage(raw)
}

func validateInstanceDomain(instanceDomain string) error {
	if strings.TrimSpace(instanceDomain) == "" {
		return fmt.Errorf("missing %s (expected a domain like %q)", EnvSoulInstanceDomain, "simulacrum.greater.website")
	}
	if strings.Contains(instanceDomain, "://") {
		return fmt.Errorf("invalid %s=%q (expected a domain without scheme)", EnvSoulInstanceDomain, instanceDomain)
	}
	if strings.ContainsAny(instanceDomain, "/ \t\r\n") {
		return fmt.Errorf("invalid %s=%q (unexpected whitespace or '/')", EnvSoulInstanceDomain, instanceDomain)
	}
	return nil
}

func InstanceDomainFromEnv() (string, error) {
	instanceDomain := strings.TrimSpace(os.Getenv(EnvSoulInstanceDomain))
	if err := validateInstanceDomain(instanceDomain); err != nil {
		return "", err
	}
	return instanceDomain, nil
}

const SSMRootPath = "/soul"

func SSMBasePath(instanceDomain string) (string, error) {
	if err := validateInstanceDomain(instanceDomain); err != nil {
		return "", err
	}
	return path.Join(SSMRootPath, instanceDomain), nil
}

func SSMPath(instanceDomain string, parts ...string) (string, error) {
	base, err := SSMBasePath(instanceDomain)
	if err != nil {
		return "", err
	}
	cleanParts := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		cleanParts = append(cleanParts, p)
	}
	return path.Join(append([]string{base}, cleanParts...)...), nil
}

func InferenceURLSSMPath(instanceDomain string) (string, error) {
	return SSMPath(instanceDomain, "inference", "url")
}

func InferenceKeySSMPath(instanceDomain string) (string, error) {
	return SSMPath(instanceDomain, "inference", "key")
}

func LesserHostInstanceKeySSMPath(instanceDomain string) (string, error) {
	return SSMPath(instanceDomain, "lesser-host", "instance-key")
}
