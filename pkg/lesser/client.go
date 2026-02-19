package lesser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	graphqlURL  string
	bearerToken string
	httpClient  *http.Client
}

type Option func(*Client)

func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.bearerToken = strings.TrimSpace(token)
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func NewClient(graphqlURL string, opts ...Option) (*Client, error) {
	graphqlURL = strings.TrimSpace(graphqlURL)
	if graphqlURL == "" {
		return nil, fmt.Errorf("missing graphqlURL")
	}
	parsed, err := url.Parse(graphqlURL)
	if err != nil {
		return nil, fmt.Errorf("parse graphqlURL: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, fmt.Errorf("invalid graphqlURL scheme %q (expected http or https)", parsed.Scheme)
	}

	c := &Client{
		graphqlURL: graphqlURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c, nil
}

type graphQLRequest struct {
	Query     string `json:"query"`
	Variables any    `json:"variables,omitempty"`
}

type graphQLResponse[T any] struct {
	Data   T              `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

func (e GraphQLError) Error() string {
	return e.Message
}

func (c *Client) doGraphQL(ctx context.Context, query string, variables any, out any) error {
	body, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("marshal graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.graphqlURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("graphql request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return fmt.Errorf("read graphql response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("graphql http %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("unmarshal graphql response: %w", err)
	}
	return nil
}
