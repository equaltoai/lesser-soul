package lesser

import (
	"context"
	"fmt"
)

const createNoteMutation = `mutation CreateNote($input: CreateNoteInput!) {
  createNote(input: $input) {
    object {
      id
      content
      createdAt
    }
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
	Actor        *Actor `json:"actor,omitempty"`
}

func (c *Client) CreateNote(ctx context.Context, input CreateNoteInput) (*Note, error) {
	vars := struct {
		Input CreateNoteInput `json:"input"`
	}{Input: input}

	var resp graphQLResponse[struct {
		CreateNote *struct {
			Object *Note `json:"object"`
		} `json:"createNote"`
	}]

	if err := c.doGraphQL(ctx, createNoteMutation, vars, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", resp.Errors[0].Message)
	}
	if resp.Data.CreateNote == nil || resp.Data.CreateNote.Object == nil {
		return nil, fmt.Errorf("graphql: missing createNote response")
	}

	note := resp.Data.CreateNote.Object
	if note.AttributedTo == nil {
		note.AttributedTo = note.Actor
	}
	return note, nil
}
