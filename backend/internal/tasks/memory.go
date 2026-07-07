package tasks

import (
	"context"
	"sort"
	"sync"
	"time"
)

type MemoryStore struct {
	mu    sync.Mutex
	tasks map[string]Task
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tasks: make(map[string]Task)}
}

func (s *MemoryStore) Enqueue(ctx context.Context, serverID string, taskName string, createdBy string) (Task, error) {
	return s.EnqueueWithPayload(ctx, serverID, taskName, TaskPayload{}, createdBy)
}

func (s *MemoryStore) EnqueueWithPayload(ctx context.Context, serverID string, taskName string, payload TaskPayload, createdBy string) (Task, error) {
	select {
	case <-ctx.Done():
		return Task{}, ctx.Err()
	default:
	}
	id, err := newTaskID()
	if err != nil {
		return Task{}, err
	}
	task := Task{
		ID: id, ServerID: serverID, Name: taskName, Payload: payload, Status: StatusPending, CreatedAt: time.Now().UTC(),
		CreatedBy:  createdBy,
		MaxRetries: 3,
		Timeout:    300,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[id] = task
	return task, nil
}

func (s *MemoryStore) ClaimPending(ctx context.Context, serverID string, limit int) ([]Task, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if limit <= 0 {
		limit = 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	pending := make([]Task, 0)
	for _, task := range s.tasks {
		if task.ServerID == serverID && task.Status == StatusPending {
			pending = append(pending, task)
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].CreatedAt.Before(pending[j].CreatedAt) })
	if len(pending) > limit {
		pending = pending[:limit]
	}
	now := time.Now().UTC()
	for i := range pending {
		pending[i].Status = StatusRunning
		pending[i].ClaimedAt = &now
		s.tasks[pending[i].ID] = pending[i]
	}
	return pending, nil
}

func (s *MemoryStore) Complete(ctx context.Context, taskID string, result TaskResult) (Task, error) {
	select {
	case <-ctx.Done():
		return Task{}, ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return Task{}, ErrNotFound{ID: taskID}
	}
	if task.Status != StatusRunning {
		return Task{}, ErrInvalidState{ID: taskID, Status: task.Status}
	}
	now := time.Now().UTC()
	task.CompletedAt = &now
	task.Result = &result
	if result.ExitCode == 0 && result.Error == "" {
		task.Status = StatusCompleted
	} else {
		task.Status = StatusFailed
	}
	s.tasks[taskID] = task
	return task, nil
}

func (s *MemoryStore) Cancel(ctx context.Context, taskID string, reason string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return Task{}, ErrNotFound{ID: taskID}
	}
	if task.Status != StatusPending && task.Status != StatusRunning {
		return Task{}, ErrInvalidState{ID: taskID, Status: task.Status}
	}
	now := time.Now().UTC()
	task.Status = StatusCanceled
	task.CompletedAt = &now
	task.Result = &TaskResult{Error: reason}
	s.tasks[taskID] = task
	return task, nil
}

func (s *MemoryStore) Get(ctx context.Context, taskID string) (Task, error) {
	select {
	case <-ctx.Done():
		return Task{}, ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return Task{}, ErrNotFound{ID: taskID}
	}
	return task, nil
}

func (s *MemoryStore) List(ctx context.Context, limit int) ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		all = append(all, t)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}
