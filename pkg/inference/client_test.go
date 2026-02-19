package inference

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Complete_MarshalsRequestAndParsesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s want=%s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path=%s want=%s", r.URL.Path, "/v1/chat/completions")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization=%q want=%q", got, "Bearer test-key")
		}

		var req openAIChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "gpt-test" {
			t.Fatalf("model=%q want=%q", req.Model, "gpt-test")
		}
		if len(req.Messages) != 2 {
			t.Fatalf("messages len=%d want=2", len(req.Messages))
		}
		if req.Messages[0].Role != "system" || req.Messages[0].Content != "sys" {
			t.Fatalf("system message=%+v", req.Messages[0])
		}
		if req.Messages[1].Role != "user" || req.Messages[1].Content != "hi" {
			t.Fatalf("user message=%+v", req.Messages[1])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "choices": [
    { "message": { "role": "assistant", "content": "hello back" } }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 5,
    "total_tokens": 15
  }
}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/v1", "test-key")
	if err != nil {
		t.Fatalf("NewClient() err=%v", err)
	}

	resp, err := client.Complete(context.Background(), CompletionRequest{
		Model:        "gpt-test",
		SystemPrompt: "sys",
		Messages: []Message{
			{Role: "user", Content: "hi"},
		},
		MaxTokens:   123,
		Temperature: 0.1,
	})
	if err != nil {
		t.Fatalf("Complete() err=%v", err)
	}
	if resp.Content != "hello back" {
		t.Fatalf("content=%q want=%q", resp.Content, "hello back")
	}
	if resp.Usage.TotalTokens != 15 {
		t.Fatalf("usage.total=%d want=15", resp.Usage.TotalTokens)
	}
}

func TestClient_Complete_FailsClosedOnMissingUsage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "choices": [
    { "message": { "role": "assistant", "content": "hello back" } }
  ]
}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, "test-key")
	if err != nil {
		t.Fatalf("NewClient() err=%v", err)
	}

	_, err = client.Complete(context.Background(), CompletionRequest{
		Model:    "gpt-test",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatalf("Complete() expected error")
	}
	if got := err.Error(); got != "inference provider non-compliant: missing usage in response" {
		t.Fatalf("err=%q", got)
	}
}
