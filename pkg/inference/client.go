package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type InferenceClient interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Stream(ctx context.Context, req CompletionRequest) (<-chan CompletionChunk, error)
}

type Client struct {
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	requireUsage bool
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithRequireUsage(requireUsage bool) Option {
	return func(c *Client) {
		c.requireUsage = requireUsage
	}
}

func NewClient(baseURL string, apiKey string, opts ...Option) (*Client, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("missing baseURL")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse baseURL: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, fmt.Errorf("invalid baseURL scheme %q (expected http or https)", parsed.Scheme)
	}

	c := &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		apiKey:       strings.TrimSpace(apiKey),
		httpClient:   &http.Client{Timeout: 60 * time.Second},
		requireUsage: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c, nil
}

type CompletionRequest struct {
	Model        string
	SystemPrompt string
	Messages     []Message
	Tools        []Tool
	MaxTokens    int
	Temperature  float32
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type CompletionResponse struct {
	Content string
	Usage   Usage
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CompletionChunk struct {
	Delta string
	Done  bool
}

var ErrStreamingNotImplemented = errors.New("streaming not implemented")

func (c *Client) Stream(_ context.Context, _ CompletionRequest) (<-chan CompletionChunk, error) {
	return nil, ErrStreamingNotImplemented
}

func (c *Client) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if strings.TrimSpace(req.Model) == "" {
		return nil, fmt.Errorf("missing model")
	}

	messages := make([]Message, 0, 1+len(req.Messages))
	if strings.TrimSpace(req.SystemPrompt) != "" {
		messages = append(messages, Message{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, req.Messages...)

	body := openAIChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Tools:       req.Tools,
	}

	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawBody))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("inference http %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var decoded openAIChatCompletionResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(decoded.Choices) == 0 || decoded.Choices[0].Message == nil {
		return nil, fmt.Errorf("inference response missing choices[0].message")
	}
	if decoded.Usage == nil && c.requireUsage {
		return nil, fmt.Errorf("inference provider non-compliant: missing usage in response")
	}

	out := &CompletionResponse{
		Content: decoded.Choices[0].Message.Content,
	}
	if decoded.Usage != nil {
		out.Usage = *decoded.Usage
	}
	return out, nil
}

type openAIChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
}

type openAIChatCompletionResponse struct {
	Choices []struct {
		Message *Message `json:"message"`
	} `json:"choices"`
	Usage *Usage `json:"usage"`
}
