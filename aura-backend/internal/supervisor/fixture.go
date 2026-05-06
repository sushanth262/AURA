package supervisor

import (
	"bytes"
	"fmt"

	"github.com/sushanth262/AURA/aura-backend/internal/fixturesdata"
	"gopkg.in/yaml.v3"
)

type FixtureScenario struct {
	ScenarioID         string         `yaml:"scenario_id"`
	DisplayTitle       string         `yaml:"display_title"`
	Severity           string         `yaml:"severity"`
	IncidentDisplayKey string         `yaml:"incident_display_key"`
	Scope              map[string]any `yaml:"scope"`
	TimelineLabels     []string       `yaml:"timeline_labels"`
	WireframeRef       string         `yaml:"wireframe_ref"`
}

func LoadFixture(name string) (*FixtureScenario, error) {
	b, err := fixturesdata.FS.ReadFile(name + ".yaml")
	if err != nil {
		return nil, err
	}
	var f FixtureScenario
	body := bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	if err := yaml.Unmarshal(body, &f); err != nil {
		return nil, err
	}
	if f.ScenarioID == "" {
		return nil, fmt.Errorf("fixture %q missing scenario_id", name)
	}
	return &f, nil
}

func DefaultFixtureName() string {
	return "inc2847_api_gateway"
}
