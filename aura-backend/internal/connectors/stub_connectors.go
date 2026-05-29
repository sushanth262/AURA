package connectors

// Stub connector constructors (Phase 5a) — YAML fixture payloads per connector id.

// NewGitHubDriver returns the GitHub fixture connector.
func NewGitHubDriver(loader FixtureLoader) Driver { return NewFixtureDriver("github", loader) }

// NewJiraDriver returns the Jira fixture connector.
func NewJiraDriver(loader FixtureLoader) Driver { return NewFixtureDriver("jira", loader) }

// NewSlackDriver returns the Slack fixture connector.
func NewSlackDriver(loader FixtureLoader) Driver { return NewFixtureDriver("slack", loader) }

// NewTeamsDriver returns the Teams fixture connector.
func NewTeamsDriver(loader FixtureLoader) Driver { return NewFixtureDriver("teams", loader) }

// NewEmailDriver returns the email fixture connector.
func NewEmailDriver(loader FixtureLoader) Driver { return NewFixtureDriver("email", loader) }
