package lesser

import (
	"context"
	"fmt"
)

const agentMemorySearchQuery = `query AgentMemorySearch(
  $query: String!
  $tags: [String!]
  $dateRange: DateRangeInput
) {
  agentMemorySearch(
    query: $query
    tags: $tags
    dateRange: $dateRange
    first: 10
  ) {
    edges {
      node {
        ... on Note {
          id
          content
          createdAt
          attributedTo { username }
        }
      }
    }
  }
}`

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
	vars := struct {
		Query     string     `json:"query"`
		Tags      []string   `json:"tags,omitempty"`
		DateRange *DateRange `json:"dateRange,omitempty"`
	}{
		Query:     params.Query,
		Tags:      params.Tags,
		DateRange: params.DateRange,
	}

	var resp graphQLResponse[struct {
		AgentMemorySearch struct {
			Edges []struct {
				Node *Note `json:"node"`
			} `json:"edges"`
		} `json:"agentMemorySearch"`
	}]

	if err := c.doGraphQL(ctx, agentMemorySearchQuery, vars, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", resp.Errors[0].Message)
	}

	out := &AgentMemorySearchResult{Notes: make([]Note, 0, len(resp.Data.AgentMemorySearch.Edges))}
	for _, edge := range resp.Data.AgentMemorySearch.Edges {
		if edge.Node == nil {
			continue
		}
		out.Notes = append(out.Notes, *edge.Node)
	}
	return out, nil
}
