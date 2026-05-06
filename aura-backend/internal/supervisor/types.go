package supervisor

import "time"

type IncidentSubmission struct {
	Title       string             `json:"title"`
	Severity    string             `json:"severity"`
	Scope       map[string]any     `json:"scope"`
	TimeWindow  map[string]any     `json:"time_window"`
	Symptoms    string             `json:"symptoms"`
	Artifacts   []any              `json:"artifacts,omitempty"`
	ScenarioKey string             `json:"scenario_key,omitempty"`
}

type QueuedResponse struct {
	TaskID     string `json:"task_id"`
	IncidentID string `json:"incident_id"`
	Status     string `json:"status"`
}

type IncidentStateResponse struct {
	IncidentID string         `json:"incident_id"`
	TaskID     string         `json:"task_id"`
	Status     string         `json:"status"`
	Severity   string         `json:"severity"`
	Title      string         `json:"title"`
	Scope      map[string]any `json:"scope,omitempty"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

type HistoryPage struct {
	Items    []IncidentSummary `json:"items"`
	Page     int               `json:"page"`
	PerPage  int               `json:"per_page"`
	Total    int               `json:"total"`
	Stats    map[string]any    `json:"stats"`
}

type IncidentSummary struct {
	IncidentID string         `json:"incident_id"`
	Title      string         `json:"title"`
	Severity   string         `json:"severity"`
	Status     string         `json:"status"`
	Scope      map[string]any `json:"scope,omitempty"`
	CreatedAt  string         `json:"created_at"`
	ResolvedAt *string        `json:"resolved_at,omitempty"`
}

type Investigation struct {
	Fixture     *FixtureScenario
	// FixtureKey is the fixturesdata bundle name without ".yaml" (e.g. inc2847_api_gateway); used for worker snapshots.
	FixtureKey string
	IncidentID string
	TaskID     string
	Title      string
	Severity   string
	Scope      map[string]any
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
