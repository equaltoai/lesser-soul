package models

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type TaskStatus string

const (
	TaskStatusRunning TaskStatus = "RUNNING"
	TaskStatusDone    TaskStatus = "DONE"
	TaskStatusFailed  TaskStatus = "FAILED"
)

type SubTaskStatus string

const (
	SubTaskStatusQueued  SubTaskStatus = "QUEUED"
	SubTaskStatusRunning SubTaskStatus = "RUNNING"
	SubTaskStatusDone    SubTaskStatus = "DONE"
	SubTaskStatusFailed  SubTaskStatus = "FAILED"
)

type AgentType string

const (
	AgentTypeResearcher AgentType = "RESEARCHER"
	AgentTypeAssistant  AgentType = "ASSISTANT"
	AgentTypeCurator    AgentType = "CURATOR"

	AgentTypeCustomCoder      AgentType = "CUSTOM:coder"
	AgentTypeCustomSummarizer AgentType = "CUSTOM:summarizer"
)

type Task struct {
	ID             string     `json:"id"`
	SK             string     `json:"sk"`
	InstanceDomain string     `json:"instance_domain"`
	Status         TaskStatus `json:"status"`
	CreatedAt      string     `json:"created_at"`
	Goal           string     `json:"goal"`
	RequestorID    string     `json:"requestor_id"`
	TotalSubtasks  int        `json:"total_subtasks"`
	DoneSubtasks   int        `json:"done_subtasks"`
	FailedSubtasks int        `json:"failed_subtasks"`
	TTL            int64      `json:"ttl"`
}

type SubTask struct {
	TaskID       string        `json:"task_id"`
	SK           string        `json:"sk"`
	AgentType    AgentType     `json:"agent_type"`
	Status       SubTaskStatus `json:"status"`
	Goal         string        `json:"goal"`
	QueueURL     string        `json:"queue_url"`
	LesserNoteID string        `json:"lesser_note_id,omitempty"`
	TokensIn     int           `json:"tokens_in"`
	TokensOut    int           `json:"tokens_out"`
	StartedAt    string        `json:"started_at,omitempty"`
	CompletedAt  string        `json:"completed_at,omitempty"`
	Error        string        `json:"error,omitempty"`
	TTL          int64         `json:"ttl"`
}

func NewTaskID() string {
	return "TASK#" + ulid.Make().String()
}

func NewSubTaskSK() string {
	return "SUB#" + ulid.Make().String()
}

func TTL30Days(now time.Time) int64 {
	return now.Add(30 * 24 * time.Hour).Unix()
}
