package notifications

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"backend/internal/incidents"
)

func TestHandlePayloadSendsIncidentCreatedNotification(t *testing.T) {
	sender := &fakeSender{}
	service := &Service{sender: sender}
	incident := incidents.Incident{
		ID:          "incident-1",
		ServerID:    "home-mini",
		ServiceName: "nginx",
		Status:      "open",
		Severity:    "critical",
		Title:       "nginx crashed",
		Summary:     "process exited with code 1",
		CreatedAt:   time.Now().UTC(),
	}
	data, _ := json.Marshal(incident)
	payload, _ := json.Marshal(map[string]json.RawMessage{
		"type": json.RawMessage(`"incident.created"`),
		"data": data,
	})

	if err := service.handlePayload(context.Background(), payload); err != nil {
		t.Fatalf("handlePayload() error = %v", err)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("messages = %#v", sender.messages)
	}
	if !strings.Contains(sender.messages[0], "New incident: nginx crashed") {
		t.Fatalf("message = %q", sender.messages[0])
	}
}

func TestHandlePayloadSkipsNonIncidentNotificationEvents(t *testing.T) {
	sender := &fakeSender{}
	service := &Service{sender: sender}
	payload := []byte(`{"type":"incident.updated","data":{"id":"incident-1"}}`)

	if err := service.handlePayload(context.Background(), payload); err != nil {
		t.Fatalf("handlePayload() error = %v", err)
	}
	if len(sender.messages) != 0 {
		t.Fatalf("messages = %#v", sender.messages)
	}
}

type fakeSender struct {
	messages []string
}

func (s *fakeSender) Send(ctx context.Context, text string) error {
	s.messages = append(s.messages, text)
	return nil
}
