package alerts

import (
	"encoding/json"
	"testing"
	"time"

	"backend/internal/domain"
)

func TestEvaluatorCreatesAlertsFromSnapshot(t *testing.T) {
	dns, _ := json.Marshal([]map[string]any{{"name": "main", "domain": "example.com", "matches_public_ip": false}})
	ports, _ := json.Marshal([]map[string]any{{"name": "web", "address": "127.0.0.1:8080", "reachable": false, "error": "connection refused"}})
	state := domain.ServerState{
		Summary: domain.ServerSummary{ID: "devbox"},
		Snapshot: domain.AgentSnapshot{
			Network: domain.NetworkSnapshot{PublicIP: "203.0.113.10", DNS: dns, Ports: ports},
			Events:  []domain.AgentEvent{{Type: "process.down", Severity: "critical", Subject: "nginx", Message: "critical process is not running", Timestamp: time.Now()}},
		},
	}
	alerts := NewEvaluator().Evaluate(state)
	if len(alerts) != 3 {
		t.Fatalf("alerts = %#v", alerts)
	}
	wantTypes := map[string]bool{"process.down": false, "dns.mismatch": false, "port.unreachable": false}
	for _, alert := range alerts {
		wantTypes[alert.Type] = true
	}
	for typ, seen := range wantTypes {
		if !seen {
			t.Fatalf("missing alert type %s in %#v", typ, alerts)
		}
	}
}
