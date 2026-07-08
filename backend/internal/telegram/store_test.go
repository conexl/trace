package telegram

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryStoreClaimLinkCreatesRecipient(t *testing.T) {
	store := NewMemoryStore()
	link, err := store.CreateLink(context.Background(), "owner@example.com", time.Minute)
	if err != nil {
		t.Fatalf("CreateLink() error = %v", err)
	}

	claimed, err := store.ClaimLink(context.Background(), link.Token, Chat{ID: 123, Type: "private", Username: "owner"})
	if err != nil {
		t.Fatalf("ClaimLink() error = %v", err)
	}
	if claimed.UserEmail != "owner@example.com" || claimed.Chat == nil || claimed.Chat.ID != 123 {
		t.Fatalf("claimed = %#v", claimed)
	}

	recipient, err := store.GetRecipient(context.Background(), "owner@example.com")
	if err != nil {
		t.Fatalf("GetRecipient() error = %v", err)
	}
	if recipient.Chat.ID != 123 {
		t.Fatalf("recipient = %#v", recipient)
	}
}

func TestMemoryStoreClaimLinkRejectsReuse(t *testing.T) {
	store := NewMemoryStore()
	link, err := store.CreateLink(context.Background(), "owner@example.com", time.Minute)
	if err != nil {
		t.Fatalf("CreateLink() error = %v", err)
	}
	if _, err := store.ClaimLink(context.Background(), link.Token, Chat{ID: 123}); err != nil {
		t.Fatalf("first ClaimLink() error = %v", err)
	}
	if _, err := store.ClaimLink(context.Background(), link.Token, Chat{ID: 456}); !errors.Is(err, ErrUsed) {
		t.Fatalf("second ClaimLink() error = %v", err)
	}
}

func TestMemoryStoreClaimLinkRejectsExpired(t *testing.T) {
	store := NewMemoryStore()
	link, err := store.CreateLink(context.Background(), "owner@example.com", -time.Second)
	if err != nil {
		t.Fatalf("CreateLink() error = %v", err)
	}
	if _, err := store.ClaimLink(context.Background(), link.Token, Chat{ID: 123}); !errors.Is(err, ErrExpired) {
		t.Fatalf("ClaimLink() error = %v", err)
	}
}
