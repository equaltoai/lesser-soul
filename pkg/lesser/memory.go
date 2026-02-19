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
)

type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type AgentMemorySearchParams struct {
	Query     string
	Tags      []string
	DateRange *DateRange
}

type AgentMemorySearchResult struct {
	Notes []Note
}

func (c *Client) AgentMemorySearch(ctx context.Context, params AgentMemorySearchParams) (*AgentMemorySearchResult, error) {
	reqBody := struct {
		Query     string   `json:"query,omitempty"`
		Tags      []string `json:"tags,omitempty"`
		DateRange *struct {
			Start string `json:"start,omitempty"`
			End   string `json:"end,omitempty"`
		} `json:"date_range,omitempty"`
		Limit int `json:"limit,omitempty"`
	}{}
	reqBody.Query = strings.TrimSpace(params.Query)
	reqBody.Tags = params.Tags
	reqBody.Limit = 10
	if params.DateRange != nil {
		reqBody.DateRange = &struct {
			Start string `json:"start,omitempty"`
			End   string `json:"end,omitempty"`
		}{
			Start: normalizeDateOnly(params.DateRange.Start),
			End:   normalizeDateOnly(params.DateRange.End),
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal agent memory search request: %w", err)
	}

	endpoint, err := c.apiURL("/api/v1/agents/memory/search")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("agent memory search request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("read agent memory search response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("agent memory search http %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var decoded struct {
		Results []struct {
			Status *struct {
				ID        string `json:"id"`
				Content   string `json:"content"`
				CreatedAt string `json:"created_at"`
				URL       string `json:"url"`
				Account   struct {
					Username string `json:"username"`
				} `json:"account"`
			} `json:"status"`
		} `json:"results"`
	}
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return nil, fmt.Errorf("unmarshal agent memory search response: %w", err)
	}

	out := &AgentMemorySearchResult{Notes: make([]Note, 0, len(decoded.Results))}
	for _, r := range decoded.Results {
		if r.Status == nil {
			continue
		}

		note := Note{
			ID:        strings.TrimSpace(r.Status.ID),
			Content:   r.Status.Content,
			CreatedAt: r.Status.CreatedAt,
			URL:       r.Status.URL,
		}
		if username := strings.TrimSpace(r.Status.Account.Username); username != "" {
			note.AttributedTo = &Actor{Username: username}
		}
		out.Notes = append(out.Notes, note)
	}
	return out, nil
}

func (c *Client) apiURL(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("invalid api path %q", path)
	}

	parsed, err := url.Parse(c.graphqlURL)
	if err != nil {
		return "", fmt.Errorf("parse graphqlURL: %w", err)
	}
	parsed.Path = path
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func normalizeDateOnly(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 10 {
		return v[:10]
	}
	return v
}
