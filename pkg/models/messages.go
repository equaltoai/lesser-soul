package models

type SubTaskQueueMessage struct {
	TaskID    string    `json:"task_id"`
	SubTaskSK string    `json:"subtask_sk"`
	AgentType AgentType `json:"agent_type"`
}

type SubTaskResultMessage struct {
	TaskID       string `json:"task_id"`
	SubTaskSK    string `json:"subtask_sk"`
	LesserNoteID string `json:"lesser_note_id,omitempty"`
	TokensIn     int    `json:"tokens_in,omitempty"`
	TokensOut    int    `json:"tokens_out,omitempty"`
	Error        string `json:"error,omitempty"`
}
