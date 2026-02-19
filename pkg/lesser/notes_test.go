package lesser

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateNote_MarshalsRequestAndParsesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s want=%s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/api/graphql" {
			t.Fatalf("path=%s want=/api/graphql", r.URL.Path)
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

		inputAny, ok := req.Variables["input"].(map[string]any)
		if !ok {
			t.Fatalf("variables.input=%T want=map[string]any", req.Variables["input"])
		}
		if inputAny["content"] != "note content" {
			t.Fatalf("input.content=%v want=%q", inputAny["content"], "note content")
		}
		if inputAny["visibility"] != "PUBLIC" {
			t.Fatalf("input.visibility=%v want=%q", inputAny["visibility"], "PUBLIC")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "data": {
    "createNote": {
      "object": {
        "id": "note_2",
        "content": "note content",
        "createdAt": "2026-01-01T00:00:00Z"
      }
    }
  }
}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL+"/api/graphql", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() err=%v", err)
	}

	got, err := client.CreateNote(context.Background(), CreateNoteInput{
		Content:    "note content",
		Visibility: VisibilityPublic,
	})
	if err != nil {
		t.Fatalf("CreateNote() err=%v", err)
	}
	if got.ID != "note_2" {
		t.Fatalf("id=%q want=%q", got.ID, "note_2")
	}
	if got.Content != "note content" {
		t.Fatalf("content=%q want=%q", got.Content, "note content")
	}
}
