package eval

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Dataset struct {
	Thresholds Thresholds `yaml:"thresholds"`
	Questions  []Question `yaml:"questions"`
}

type Thresholds struct {
	MetadataHitRate          float64 `yaml:"metadata_hit_rate"`
	TopSourceMetadataHitRate float64 `yaml:"top_source_metadata_hit_rate"`
	MinSourcesPerQuestion    int     `yaml:"min_sources_per_question"`
}

type Question struct {
	ID                      string           `yaml:"id"`
	Question                string           `yaml:"question"`
	TopK                    int              `yaml:"top_k"`
	Preferences             Preferences      `yaml:"preferences"`
	ExpectedMetadata        ExpectedMetadata `yaml:"expected_metadata"`
	ExpectedPreviewContains []string         `yaml:"expected_preview_contains"`
}

type Preferences struct {
	Layers        []string `yaml:"layers" json:"layers,omitempty"`
	Categories    []string `yaml:"categories" json:"categories,omitempty"`
	Platforms     []string `yaml:"platforms" json:"platforms,omitempty"`
	SourceKinds   []string `yaml:"source_kinds" json:"source_kinds,omitempty"`
	SourceQuality []string `yaml:"source_quality" json:"source_quality,omitempty"`
}

type ExpectedMetadata struct {
	Layers        []string `yaml:"layers"`
	Categories    []string `yaml:"categories"`
	Platforms     []string `yaml:"platforms"`
	SourceKinds   []string `yaml:"source_kinds"`
	SourceQuality []string `yaml:"source_quality"`
}

func LoadDataset(path string) (*Dataset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read dataset: %w", err)
	}

	var dataset Dataset
	if err := yaml.Unmarshal(data, &dataset); err != nil {
		return nil, fmt.Errorf("failed to parse dataset yaml: %w", err)
	}
	if err := dataset.Validate(); err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (d *Dataset) Validate() error {
	if len(d.Questions) == 0 {
		return fmt.Errorf("dataset must contain at least one question")
	}
	if d.Thresholds.MinSourcesPerQuestion <= 0 {
		d.Thresholds.MinSourcesPerQuestion = 1
	}
	seen := make(map[string]struct{}, len(d.Questions))
	for _, question := range d.Questions {
		if question.ID == "" {
			return fmt.Errorf("question id is required")
		}
		if _, exists := seen[question.ID]; exists {
			return fmt.Errorf("duplicate question id %q", question.ID)
		}
		seen[question.ID] = struct{}{}
		if question.Question == "" {
			return fmt.Errorf("question %q text is required", question.ID)
		}
		if question.TopK <= 0 {
			return fmt.Errorf("question %q top_k must be greater than zero", question.ID)
		}
	}
	return nil
}
