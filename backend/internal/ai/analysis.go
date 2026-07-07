package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"backend/internal/domain"
	"backend/internal/incidents"

	"go.uber.org/zap"
)

// IncidentAnalysis represents structured AI analysis
type IncidentAnalysis struct {
	Summary     string   `json:"summary"`
	RootCause   string   `json:"root_cause"`
	Severity    string   `json:"severity"`
	Suggestions []string `json:"suggestions"`
	Confidence  float64  `json:"confidence"`
}

// IncidentContext gives the model enough surrounding telemetry to diagnose
// without granting it permission to execute remediation actions.
type IncidentContext struct {
	Server  *domain.ServerState `json:"server,omitempty"`
	Metrics []domain.Metric     `json:"metrics,omitempty"`
}

// Analyzer provides AI-powered incident analysis
type Analyzer struct {
	client *Client
	logger *zap.Logger
}

// NewAnalyzer creates new analyzer
func NewAnalyzer(client *Client, logger *zap.Logger) *Analyzer {
	return &Analyzer{
		client: client,
		logger: logger.Named("analyzer"),
	}
}

// AnalyzeIncident analyzes incident and returns structured insights
func (a *Analyzer) AnalyzeIncident(ctx context.Context, incident *incidents.Incident, extra ...IncidentContext) (*IncidentAnalysis, error) {
	if !a.client.IsEnabled() {
		return &IncidentAnalysis{
			Summary:     "AI analysis unavailable. Configure AI_API_KEY to enable.",
			RootCause:   "Not analyzed",
			Severity:    incident.Severity,
			Suggestions: []string{"Configure AI_API_KEY environment variable to enable AI-powered analysis"},
			Confidence:  0,
		}, nil
	}

	systemPrompt := `You are an expert DevOps engineer analyzing service incidents. Provide structured JSON analysis.

Always respond with valid JSON containing:
- summary: 1-2 sentence incident summary
- root_cause: most likely cause based on available data
- severity: "critical", "warning", or "info"
- suggestions: array of 2-4 safe remediation steps that an operator can approve manually
- confidence: 0.0-1.0 score of your confidence

Be concise and actionable. Do not claim that you executed any command. Treat restart, watchdog changes, diagnostics, and rollback as operator-approved actions only.`

	var incidentContext *IncidentContext
	if len(extra) > 0 {
		incidentContext = &extra[0]
	}
	userPrompt := a.buildIncidentPrompt(incident, incidentContext)

	response, err := a.client.Analyze(ctx, systemPrompt, userPrompt)
	if err != nil {
		a.logger.Warn("AI analysis failed", zap.Error(err))
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	analysis, err := a.parseAnalysis(response)
	if err != nil {
		a.logger.Warn("Failed to parse AI response", zap.Error(err), zap.String("response", response))
		return &IncidentAnalysis{
			Summary:     response,
			RootCause:   "Unable to parse structured analysis",
			Severity:    incident.Severity,
			Suggestions: []string{"Review incident timeline manually"},
			Confidence:  0.5,
		}, nil
	}

	return analysis, nil
}

func (a *Analyzer) buildIncidentPrompt(incident *incidents.Incident, ctx *IncidentContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Incident: %s\n", incident.Title))
	sb.WriteString(fmt.Sprintf("Service: %s\n", incident.ServiceName))
	sb.WriteString(fmt.Sprintf("Status: %s\n", incident.Status))
	sb.WriteString(fmt.Sprintf("Severity: %s\n", incident.Severity))
	sb.WriteString(fmt.Sprintf("Created: %s\n", incident.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Summary: %s\n\n", incident.Summary))

	sb.WriteString("Timeline:\n")
	for _, event := range incident.Timeline {
		sb.WriteString(fmt.Sprintf("- [%s] %s", event.Timestamp.Format("15:04:05"), event.Title))
		if event.Message != "" {
			sb.WriteString(fmt.Sprintf(": %s", event.Message))
		}
		if event.ExitCode != 0 {
			sb.WriteString(fmt.Sprintf(" (exit code: %d)", event.ExitCode))
		}
		sb.WriteString("\n")
	}

	if ctx == nil {
		return sb.String()
	}

	if ctx.Server != nil {
		state := ctx.Server
		sb.WriteString("\nCurrent server state:\n")
		sb.WriteString(fmt.Sprintf("- Server ID: %s\n", state.Summary.ID))
		sb.WriteString(fmt.Sprintf("- Status: %s\n", state.Summary.Status))
		sb.WriteString(fmt.Sprintf("- Hostname: %s\n", state.Summary.Hostname))
		sb.WriteString(fmt.Sprintf("- Platform: %s\n", state.Summary.Platform))
		sb.WriteString(fmt.Sprintf("- CPU: %.1f%%\n", state.Summary.CPUPercent))
		sb.WriteString(fmt.Sprintf("- Memory: %.1f%%\n", state.Summary.MemoryUsed))
		sb.WriteString(fmt.Sprintf("- Public IP: %s\n", state.Summary.PublicIP))
		sb.WriteString(fmt.Sprintf("- Applied config revision: %d\n", state.Summary.AppliedConfigRevision))
		sb.WriteString(fmt.Sprintf("- Desired config revision: %d\n", state.Summary.DesiredConfigRevision))

		sb.WriteString("\nRelevant processes:\n")
		count := 0
		for _, proc := range state.Snapshot.Processes {
			if proc.Name == incident.ServiceName || proc.Service == incident.ServiceName || strings.Contains(proc.Match, incident.ServiceName) {
				sb.WriteString(fmt.Sprintf("- %s service=%s running=%v status=%s exit=%d cpu=%.1f rss=%d error=%s\n",
					proc.Name, proc.Service, proc.Running, proc.Status, proc.LastExitCode, proc.CPUPercent, proc.MemoryRSS, proc.Error))
				count++
			}
		}
		if count == 0 {
			sb.WriteString("- No matching process snapshot found\n")
		}

		sb.WriteString("\nRecent logs:\n")
		logsWritten := 0
		for _, chunk := range state.Snapshot.Logs {
			if logsWritten >= 3 {
				break
			}
			if chunk.Data == "" {
				continue
			}
			data := chunk.Data
			if len(data) > 1200 {
				data = data[len(data)-1200:]
			}
			sb.WriteString(fmt.Sprintf("- %s (%s):\n%s\n", chunk.Name, chunk.Path, data))
			logsWritten++
		}
		if logsWritten == 0 {
			sb.WriteString("- No log chunks available\n")
		}
	}

	if len(ctx.Metrics) > 0 {
		sb.WriteString("\nRecent metrics around incident:\n")
		start := 0
		if len(ctx.Metrics) > 12 {
			start = len(ctx.Metrics) - 12
		}
		for _, metric := range ctx.Metrics[start:] {
			sb.WriteString(fmt.Sprintf("- %s cpu=%.1f memory=%.1f net_in=%d net_out=%d\n",
				metric.Timestamp.Format("15:04:05"), metric.CPU, metric.Memory, metric.NetIn, metric.NetOut))
		}
	}

	return sb.String()
}

func (a *Analyzer) parseAnalysis(response string) (*IncidentAnalysis, error) {
	// Try to extract JSON from response
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var analysis IncidentAnalysis
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		return nil, err
	}

	// Validate and set defaults
	if analysis.Summary == "" {
		analysis.Summary = "Analysis completed"
	}
	if analysis.RootCause == "" {
		analysis.RootCause = "Unknown"
	}
	if analysis.Severity == "" {
		analysis.Severity = "warning"
	}
	if len(analysis.Suggestions) == 0 {
		analysis.Suggestions = []string{"Review service logs", "Check resource usage"}
	}
	if analysis.Confidence == 0 {
		analysis.Confidence = 0.7
	}

	return &analysis, nil
}
