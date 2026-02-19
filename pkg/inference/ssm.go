package inference

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type GetParametersAPI interface {
	GetParameters(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error)
}

func LoadURLAndKeyFromSSM(ctx context.Context, api GetParametersAPI, urlPath string, keyPath string) (string, string, error) {
	urlPath = strings.TrimSpace(urlPath)
	keyPath = strings.TrimSpace(keyPath)
	if urlPath == "" {
		return "", "", fmt.Errorf("missing inference url ssm path")
	}
	if keyPath == "" {
		return "", "", fmt.Errorf("missing inference key ssm path")
	}

	out, err := api.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          []string{urlPath, keyPath},
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", "", fmt.Errorf("ssm get parameters: %w", err)
	}
	if len(out.InvalidParameters) > 0 {
		return "", "", fmt.Errorf("ssm missing parameters: %s", strings.Join(out.InvalidParameters, ", "))
	}

	var (
		rawURL string
		rawKey string
	)
	for _, p := range out.Parameters {
		name := aws.ToString(p.Name)
		switch name {
		case urlPath:
			rawURL = strings.TrimSpace(aws.ToString(p.Value))
		case keyPath:
			rawKey = strings.TrimSpace(aws.ToString(p.Value))
		}
	}

	if rawURL == "" {
		return "", "", fmt.Errorf("ssm parameter %q was empty", urlPath)
	}
	if rawKey == "" {
		return "", "", fmt.Errorf("ssm parameter %q was empty", keyPath)
	}
	return rawURL, rawKey, nil
}
