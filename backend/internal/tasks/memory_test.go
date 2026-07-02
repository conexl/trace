package tasks

import (
	"context"
	"testing"
)

func TestMemoryStoreTaskLifecycle(t *testing.T) {
	store := NewMemoryStore()
	task, err := store.Enqueue(context.Background(), "devbox", "disk-usage")
	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	claimed, err := store.ClaimPending(context.Background(), "devbox", 1)
	if err != nil {
		t.Fatalf("ClaimPending() error = %v", err)
	}
	if len(claimed) != 1 || claimed[0].ID != task.ID || claimed[0].Status != StatusRunning {
		t.Fatalf("claimed = %#v", claimed)
	}
	completed, err := store.Complete(context.Background(), task.ID, TaskResult{ExitCode: 0, Stdout: "ok"})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if completed.Status != StatusCompleted || completed.Result.Stdout != "ok" {
		t.Fatalf("completed = %#v", completed)
	}
}

func TestMemoryStoreCompleteRequiresRunning(t *testing.T) {
	store := NewMemoryStore()
	task, err := store.Enqueue(context.Background(), "devbox", "disk-usage")
	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if _, err := store.Complete(context.Background(), task.ID, TaskResult{}); err == nil {
		t.Fatal("Complete() expected invalid state error")
	}
}
