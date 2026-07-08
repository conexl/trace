package notifications

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"backend/internal/incidents"
	"backend/internal/telegram"
)

func TestHandlePayloadSendsIncidentCreatedNotificationToLinkedRecipient(t *testing.T) {
	sender := &fakeSender{}
	store := telegram.NewMemoryStore()
	link, err := store.CreateLink(context.Background(), "owner@example.com", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ClaimLink(context.Background(), link.Token, telegram.Chat{ID: 123, Type: "private", Username: "owner"}); err != nil {
		t.Fatal(err)
	}
	service := &Service{sender: sender, store: store}
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
	if sender.messages[0].chatID != "123" || !strings.Contains(sender.messages[0].text, "New incident: nginx crashed") {
		t.Fatalf("message = %#v", sender.messages[0])
	}
}

func TestHandleTelegramUpdateClaimsStartToken(t *testing.T) {
	sender := &fakeSender{}
	store := telegram.NewMemoryStore()
	link, err := store.CreateLink(context.Background(), "owner@example.com", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	service := &Service{sender: sender, store: store}

	err = service.handleTelegramUpdate(context.Background(), TelegramUpdate{
		UpdateID: 1,
		Message: &TelegramMessage{
			Text: "/start " + link.Token,
			Chat: telegram.Chat{ID: 456, Type: "private", Username: "owner"},
		},
	})
	if err != nil {
		t.Fatalf("handleTelegramUpdate() error = %v", err)
	}
	recipient, err := store.GetRecipient(context.Background(), "owner@example.com")
	if err != nil {
		t.Fatalf("GetRecipient() error = %v", err)
	}
	if recipient.Chat.ID != 456 {
		t.Fatalf("recipient = %#v", recipient)
	}
	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].text, "connected") {
		t.Fatalf("messages = %#v", sender.messages)
	}
}

func TestHandleTelegramUpdateAcceptsBotMentionStartCommand(t *testing.T) {
	sender := &fakeSender{}
	store := telegram.NewMemoryStore()
	link, err := store.CreateLink(context.Background(), "owner@example.com", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	service := &Service{sender: sender, store: store}

	err = service.handleTelegramUpdate(context.Background(), TelegramUpdate{
		UpdateID: 1,
		Message: &TelegramMessage{
			Text: "/start@TraceDemoBot " + link.Token,
			Chat: telegram.Chat{ID: 456, Type: "private", Username: "owner"},
		},
	})
	if err != nil {
		t.Fatalf("handleTelegramUpdate() error = %v", err)
	}
	if _, err := store.GetRecipient(context.Background(), "owner@example.com"); err != nil {
		t.Fatalf("GetRecipient() error = %v", err)
	}
}

func TestHandlePayloadUsesLegacyChatWhenNoLinkedRecipients(t *testing.T) {
	sender := &fakeSender{}
	service := &Service{sender: sender, store: telegram.NewMemoryStore(), legacyChatID: "999"}
	incident := incidents.Incident{ID: "incident-1", Title: "nginx crashed", ServiceName: "nginx", Severity: "critical"}
	data, _ := json.Marshal(incident)
	payload, _ := json.Marshal(map[string]json.RawMessage{
		"type": json.RawMessage(`"incident.created"`),
		"data": data,
	})

	if err := service.handlePayload(context.Background(), payload); err != nil {
		t.Fatalf("handlePayload() error = %v", err)
	}
	if len(sender.messages) != 1 || sender.messages[0].chatID != "999" {
		t.Fatalf("messages = %#v", sender.messages)
	}
}

func TestHandlePayloadSkipsNonIncidentNotificationEvents(t *testing.T) {
	sender := &fakeSender{}
	service := &Service{sender: sender, store: telegram.NewMemoryStore()}
	payload := []byte(`{"type":"incident.updated","data":{"id":"incident-1"}}`)

	if err := service.handlePayload(context.Background(), payload); err != nil {
		t.Fatalf("handlePayload() error = %v", err)
	}
	if len(sender.messages) != 0 {
		t.Fatalf("messages = %#v", sender.messages)
	}
}

type sentMessage struct {
	chatID string
	text   string
}

type fakeSender struct {
	messages []sentMessage
}

func (s *fakeSender) Send(ctx context.Context, chatID string, text string) error {
	s.messages = append(s.messages, sentMessage{chatID: chatID, text: text})
	return nil
}
