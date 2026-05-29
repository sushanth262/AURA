package workersvc

import (
	"bytes"
	"fmt"

	"github.com/sushanth262/AURA/aura-backend/internal/fixturesdata"
	"gopkg.in/yaml.v3"
)

func loadScenarioYAML(baseName string) (map[string]any, error) {
	b, err := fixturesdata.FS.ReadFile(baseName + ".yaml")
	if err != nil {
		return nil, fmt.Errorf("fixture %q: %w", baseName, err)
	}
	body := bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	var root map[string]any
	if err := yaml.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	return root, nil
}

func extractSourceMock(root map[string]any, source string) any {
	sm, _ := root["source_mocks"].(map[string]any)
	if sm == nil {
		return nil
	}
	return sm[source]
}
