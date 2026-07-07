package tasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

type Task struct {
	ID          string      `json:"id" bson:"_id"`
	ServerID    string      `json:"server_id" bson:"server_id"`
	Name        string      `json:"name" bson:"name"`
	Payload     TaskPayload `json:"payload,omitempty" bson:"payload,omitempty"`
	Status      Status      `json:"status" bson:"status"`
	CreatedAt   time.Time   `json:"created_at" bson:"created_at"`
	CreatedBy   string      `json:"created_by,omitempty" bson:"created_by,omitempty"`
	ClaimedAt   *time.Time  `json:"claimed_at,omitempty" bson:"claimed_at,omitempty"`
	CompletedAt *time.Time  `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	Result      *TaskResult `json:"result,omitempty" bson:"result,omitempty"`
	Retries     int         `json:"retries" bson:"retries"`
	MaxRetries  int         `json:"max_retries" bson:"max_retries"`
	Timeout     int         `json:"timeout_seconds,omitempty" bson:"timeout_seconds,omitempty"`
}

type TaskPayload struct {
	Service string   `json:"service,omitempty" bson:"service,omitempty"`
	Action  string   `json:"action,omitempty" bson:"action,omitempty"`
	Domains []string `json:"domains,omitempty" bson:"domains,omitempty"`
}

type TaskResult struct {
	ExitCode   int       `json:"exit_code" bson:"exit_code"`
	Stdout     string    `json:"stdout" bson:"stdout"`
	Stderr     string    `json:"stderr" bson:"stderr"`
	DurationMS int64     `json:"duration_ms" bson:"duration_ms"`
	StartedAt  time.Time `json:"started_at" bson:"started_at"`
	Error      string    `json:"error,omitempty" bson:"error,omitempty"`
}

type Store interface {
	Enqueue(ctx context.Context, serverID string, taskName string, createdBy string) (Task, error)
	EnqueueWithPayload(ctx context.Context, serverID string, taskName string, payload TaskPayload, createdBy string) (Task, error)
	ClaimPending(ctx context.Context, serverID string, limit int) ([]Task, error)
	Complete(ctx context.Context, taskID string, result TaskResult) (Task, error)
	Cancel(ctx context.Context, taskID string, reason string) (Task, error)
	Get(ctx context.Context, taskID string) (Task, error)
	List(ctx context.Context, limit int) ([]Task, error)
}

type ErrNotFound struct {
	ID string
}

func (e ErrNotFound) Error() string { return "task not found: " + e.ID }

type ErrInvalidState struct {
	ID     string
	Status Status
}

func (e ErrInvalidState) Error() string {
	return fmt.Sprintf("task %s cannot transition from %s", e.ID, e.Status)
}

func newTaskID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "tsk_" + hex.EncodeToString(raw), nil
}
