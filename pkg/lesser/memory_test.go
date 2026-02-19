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
		if r.URL.Path != "/api/v1/agents/memory/search" {
			t.Fatalf("path=%s want=/api/v1/agents/memory/search", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization=%q want=%q", got, "Bearer test-token")
		}

		var req struct {
			Query     string   `json:"query"`
			Tags      []string `json:"tags"`
			DateRange *struct {
				Start string `json:"start"`
				End   string `json:"end"`
			} `json:"date_range"`
			Limit int `json:"limit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Query != "memory query" {
			t.Fatalf("query=%v want=%q", req.Query, "memory query")
		}
		if len(req.Tags) != 2 || req.Tags[0] != "a" || req.Tags[1] != "b" {
			t.Fatalf("tags=%v want=[a b]", req.Tags)
		}
		if req.DateRange == nil || req.DateRange.Start != "2026-01-01" || req.DateRange.End != "2026-01-02" {
			t.Fatalf("date_range=%v", req.DateRange)
		}
		if req.Limit != 10 {
			t.Fatalf("limit=%d want=10", req.Limit)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "results": [
    {
      "status": {
        "id": "note_1",
        "content": "hello",
        "created_at": "2026-01-01T00:00:00Z",
        "url": "https://example.com/objects/note_1",
        "account": { "username": "soul-researcher" }
      }
    }
  ],
  "total": 1,
  "query_time_ms": 5
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
