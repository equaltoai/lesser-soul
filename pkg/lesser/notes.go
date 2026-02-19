package lesser

import (
	"context"
	"fmt"
)

const createNoteMutation = `mutation CreateNote($input: CreateNoteInput!) {
  createNote(input: $input) {
    id
    content
    createdAt
    url
  }
}`

type Visibility string

const (
	VisibilityPublic    Visibility = "PUBLIC"
	VisibilityUnlisted  Visibility = "UNLISTED"
	VisibilityFollowers Visibility = "FOLLOWERS"
	VisibilityDirect    Visibility = "DIRECT"
)

type CreateNoteInput struct {
	Content    string     `json:"content"`
	Visibility Visibility `json:"visibility"`
}

type Actor struct {
	Username string `json:"username"`
}

type Note struct {
	ID           string `json:"id"`
	Content      string `json:"content"`
	CreatedAt    string `json:"createdAt"`
	URL          string `json:"url,omitempty"`
	AttributedTo *Actor `json:"attributedTo,omitempty"`
}

func (c *Client) CreateNote(ctx context.Context, input CreateNoteInput) (*Note, error) {
	vars := struct {
		Input CreateNoteInput `json:"input"`
	}{Input: input}

	var resp graphQLResponse[struct {
		CreateNote *Note `json:"createNote"`
	}]

	if err := c.doGraphQL(ctx, createNoteMutation, vars, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", resp.Errors[0].Message)
	}
	if resp.Data.CreateNote == nil {
		return nil, fmt.Errorf("graphql: missing createNote response")
	}
	return resp.Data.CreateNote, nil
}
