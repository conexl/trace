package transport

import (
	"context"

	"agent/internal/collectors"
)

type Client interface {
	SendSnapshots(ctx context.Context, snapshots []collectors.Snapshot) error
}

type NopClient struct{}

func (NopClient) SendSnapshots(context.Context, []collectors.Snapshot) error {
	return nil
}
