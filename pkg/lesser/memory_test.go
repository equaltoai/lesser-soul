package lesser

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_AgentMemorySearch_MarshalsRequestAndParsesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s want=%s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/api/graphql" {
			t.Fatalf("path=%s want=/api/graphql", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization=%q want=%q", got, "Bearer test-token")
		}

		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Query == "" {
			t.Fatalf("query empty")
		}
		if _, ok := req.Variables["query"]; !ok {
			t.Fatalf("variables.query missing")
		}
		if req.Variables["query"] != "memory query" {
			t.Fatalf("variables.query=%v want=%q", req.Variables["query"], "memory query")
		}

		tagsAny, ok := req.Variables["tags"].([]any)
		if !ok {
			t.Fatalf("variables.tags=%T want=[]any", req.Variables["tags"])
		}
		if len(tagsAny) != 2 || tagsAny[0] != "a" || tagsAny[1] != "b" {
			t.Fatalf("variables.tags=%v want=[a b]", tagsAny)
		}

		drAny, ok := req.Variables["dateRange"].(map[string]any)
		if !ok {
			t.Fatalf("variables.dateRange=%T want=map[string]any", req.Variables["dateRange"])
		}
		if drAny["start"] != "2026-01-01T00:00:00Z" || drAny["end"] != "2026-01-02T00:00:00Z" {
			t.Fatalf("variables.dateRange=%v", drAny)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "data": {
    "agentMemorySearch": {
      "edges": [
        {
          "node": {
            "id": "note_1",
            "content": "hello",
            "createdAt": "2026-01-01T00:00:00Z",
            "attributedTo": { "username": "soul-researcher" }
          }
        }
      ]
    }
  }
}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/api/graphql", WithBearerToken("test-token"), WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() err=%v", err)
	}

	got, err := client.AgentMemorySearch(context.Background(), AgentMemorySearchParams{
		Query: "memory query",
		Tags:  []string{"a", "b"},
		DateRange: &DateRange{
			Start: "2026-01-01T00:00:00Z",
			End:   "2026-01-02T00:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("AgentMemorySearch() err=%v", err)
	}
	if len(got.Notes) != 1 {
		t.Fatalf("notes len=%d want=1", len(got.Notes))
	}
	if got.Notes[0].ID != "note_1" {
		t.Fatalf("note id=%q want=%q", got.Notes[0].ID, "note_1")
	}
	if got.Notes[0].AttributedTo == nil || got.Notes[0].AttributedTo.Username != "soul-researcher" {
		t.Fatalf("note attributedTo=%+v", got.Notes[0].AttributedTo)
	}
}
