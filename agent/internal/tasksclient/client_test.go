package tasksclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agent/internal/commands"
	"agent/internal/config"
)

func TestPollAndComplete(t *testing.T) {
	var completed bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agent/tasks":
			if r.URL.Query().Get("agent_id") != "devbox" {
				t.Fatalf("agent_id = %q", r.URL.Query().Get("agent_id"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"tasks": []Task{{ID: "tsk_1", ServerID: "devbox", Name: "disk-usage", Status: "running"}}})
		case "/v1/agent/tasks/tsk_1/result":
			completed = true
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	client, err := New(config.CloudConfig{Endpoint: server.URL, Token: "agent-token"})
	if err != nil {
		t.Fatal(err)
	}
	tasks, err := client.Poll(context.Background(), "devbox", 1)
	if err != nil {
		t.Fatalf("Poll() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "tsk_1" {
		t.Fatalf("tasks = %#v", tasks)
	}
	if err := client.Complete(context.Background(), "tsk_1", TaskResult{ExitCode: 0}); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if !completed {
		t.Fatal("complete endpoint was not called")
	}
}

func TestFromCommandResultIncludesError(t *testing.T) {
	result := FromCommandResult(commands.Result{ExitCode: 1, Duration: time.Second, StartedAt: time.Unix(1, 0)}, context.Canceled)
	if result.Error == "" || result.DurationMS != 1000 {
		t.Fatalf("result = %#v", result)
	}
}
