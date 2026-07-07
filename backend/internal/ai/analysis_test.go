package ai

import (
	"context"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/incidents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestParseAnalysis_ValidJSON(t *testing.T) {
	analyzer := NewAnalyzer(nil, zap.NewNop())

	jsonResp := `{"summary":"Service crashed due to OOM","root_cause":"Memory limit exceeded","severity":"critical","suggestions":["Increase memory limit","Check for memory leaks"],"confidence":0.85}`

	analysis, err := analyzer.parseAnalysis(jsonResp)
	require.NoError(t, err)

	assert.Equal(t, "Service crashed due to OOM", analysis.Summary)
	assert.Equal(t, "Memory limit exceeded", analysis.RootCause)
	assert.Equal(t, "critical", analysis.Severity)
	assert.Len(t, analysis.Suggestions, 2)
	assert.InDelta(t, 0.85, analysis.Confidence, 0.01)
}

func TestParseAnalysis_WithMarkdownCodeBlock(t *testing.T) {
	analyzer := NewAnalyzer(nil, zap.NewNop())

	jsonResp := "```json\n{\"summary\":\"Test\",\"root_cause\":\"Unknown\",\"severity\":\"warning\",\"suggestions\":[\"Check logs\"],\"confidence\":0.7}\n```"

	analysis, err := analyzer.parseAnalysis(jsonResp)
	require.NoError(t, err)

	assert.Equal(t, "Test", analysis.Summary)
	assert.Equal(t, "Unknown", analysis.RootCause)
}

func TestParseAnalysis_MissingFields(t *testing.T) {
	analyzer := NewAnalyzer(nil, zap.NewNop())

	// Missing fields should get defaults
	jsonResp := `{"summary":"Partial analysis"}`

	analysis, err := analyzer.parseAnalysis(jsonResp)
	require.NoError(t, err)

	assert.Equal(t, "Partial analysis", analysis.Summary)
	assert.Equal(t, "Unknown", analysis.RootCause) // default
	assert.Equal(t, "warning", analysis.Severity)  // default
	assert.Len(t, analysis.Suggestions, 2)         // default suggestions
	assert.InDelta(t, 0.7, analysis.Confidence, 0.01)
}

func TestParseAnalysis_InvalidJSON(t *testing.T) {
	analyzer := NewAnalyzer(nil, zap.NewNop())

	_, err := analyzer.parseAnalysis("not valid json")
	assert.Error(t, err)
}

func TestBuildIncidentPrompt(t *testing.T) {
	analyzer := NewAnalyzer(nil, zap.NewNop())

	incident := &incidents.Incident{
		ID:          "test:nginx:2026-07-07-120000",
		ServerID:    "test-server",
		ServiceName: "nginx",
		Status:      "open",
		Severity:    "critical",
		Title:       "nginx crashed",
		Summary:     "Service exited with code 1",
		CreatedAt:   time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC),
		Timeline: []incidents.TimelineEvent{
			{
				ID:        "event-1",
				Type:      "crash",
				Timestamp: time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC),
				Title:     "Service crashed",
				Message:   "Exited unexpectedly",
				ExitCode:  1,
			},
			{
				ID:        "event-2",
				Type:      "restart",
				Timestamp: time.Date(2026, 7, 7, 12, 0, 1, 0, time.UTC),
				Title:     "Restart attempted",
				Action:    "restart",
				Result:    "success",
			},
		},
	}

	prompt := analyzer.buildIncidentPrompt(incident, nil)

	assert.Contains(t, prompt, "Incident: nginx crashed")
	assert.Contains(t, prompt, "Service: nginx")
	assert.Contains(t, prompt, "Status: open")
	assert.Contains(t, prompt, "Severity: critical")
	assert.Contains(t, prompt, "Timeline:")
	assert.Contains(t, prompt, "Service crashed")
	assert.Contains(t, prompt, "exit code: 1")
	assert.Contains(t, prompt, "Restart attempted")
}

func TestAnalyzeIncident_Disabled(t *testing.T) {
	client := NewClient(ClientParams{
		Config: configWithAI(""), // No API key
		Logger: zap.NewNop(),
	})
	analyzer := NewAnalyzer(client, zap.NewNop())

	incident := &incidents.Incident{
		ID:          "test:service:2026-07-07-120000",
		ServiceName: "test-service",
		Severity:    "warning",
		Title:       "Test incident",
		Summary:     "Test summary",
		CreatedAt:   time.Now(),
	}

	analysis, err := analyzer.AnalyzeIncident(context.Background(), incident)
	require.NoError(t, err)

	// Should return fallback message
	assert.Contains(t, analysis.Summary, "AI analysis unavailable")
	assert.Contains(t, analysis.Suggestions[0], "Configure AI_API_KEY")
	assert.Equal(t, 0.0, analysis.Confidence)
}

func configWithAI(apiKey string) config.Config {
	return config.Config{
		AI: config.AIConfig{
			APIKey:  apiKey,
			BaseURL: "https://api.deepseek.com",
			Model:   "deepseek-chat",
		},
	}
}
