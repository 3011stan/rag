package eval

import "strings"

type Report struct {
	Results                  []QuestionResult
	QuestionsTotal           int
	MetadataHitRate          float64
	TopSourceMetadataHitRate float64
	EmptySourceCount         int
	Passed                   bool
	Thresholds               Thresholds
}

type QuestionResult struct {
	ID                      string
	SourceCount             int
	MetadataHit             bool
	TopSourceMetadataHit    bool
	PreviewHit              bool
	MissingExpectedMetadata []string
	Status                  string
}

func evaluateQuestion(question Question, response *askResponse) QuestionResult {
	result := QuestionResult{
		ID:          question.ID,
		SourceCount: len(response.Sources),
		PreviewHit:  previewHit(response.Sources, question.ExpectedPreviewContains),
		Status:      "pass",
	}

	for i, source := range response.Sources {
		missing := missingMetadata(source.Metadata, question.ExpectedMetadata)
		if len(missing) == 0 {
			result.MetadataHit = true
			if i == 0 {
				result.TopSourceMetadataHit = true
			}
			break
		}
		if i == 0 {
			result.MissingExpectedMetadata = missing
		}
	}

	if result.SourceCount == 0 || !result.MetadataHit {
		result.Status = "fail"
	}
	return result
}

func summarize(thresholds Thresholds, results []QuestionResult) *Report {
	report := &Report{
		Results:        results,
		QuestionsTotal: len(results),
		Thresholds:     thresholds,
		Passed:         true,
	}
	if report.Thresholds.MinSourcesPerQuestion <= 0 {
		report.Thresholds.MinSourcesPerQuestion = 1
	}

	var metadataHits, topHits int
	for _, result := range results {
		if result.MetadataHit {
			metadataHits++
		}
		if result.TopSourceMetadataHit {
			topHits++
		}
		if result.SourceCount < report.Thresholds.MinSourcesPerQuestion {
			report.EmptySourceCount++
		}
	}

	if len(results) > 0 {
		report.MetadataHitRate = float64(metadataHits) / float64(len(results))
		report.TopSourceMetadataHitRate = float64(topHits) / float64(len(results))
	}

	if report.MetadataHitRate < report.Thresholds.MetadataHitRate ||
		report.TopSourceMetadataHitRate < report.Thresholds.TopSourceMetadataHitRate ||
		report.EmptySourceCount > 0 {
		report.Passed = false
	}
	return report
}

func missingMetadata(metadata map[string]interface{}, expected ExpectedMetadata) []string {
	var missing []string
	if !matchesAny(metadata["layer"], expected.Layers) {
		missing = appendMissing(missing, "layer", expected.Layers)
	}
	if !matchesAny(metadata["category"], expected.Categories) {
		missing = appendMissing(missing, "category", expected.Categories)
	}
	if !matchesAny(metadata["platform"], expected.Platforms) {
		missing = appendMissing(missing, "platform", expected.Platforms)
	}
	if !matchesAny(metadata["source_kind"], expected.SourceKinds) {
		missing = appendMissing(missing, "source_kind", expected.SourceKinds)
	}
	if !matchesAny(metadata["source_quality"], expected.SourceQuality) {
		missing = appendMissing(missing, "source_quality", expected.SourceQuality)
	}
	return missing
}

func appendMissing(missing []string, field string, expected []string) []string {
	if len(expected) == 0 {
		return missing
	}
	return append(missing, field)
}

func matchesAny(value interface{}, expected []string) bool {
	if len(expected) == 0 {
		return true
	}
	text, ok := value.(string)
	if !ok {
		return false
	}
	for _, candidate := range expected {
		if text == candidate {
			return true
		}
	}
	return false
}

func previewHit(sources []SourceInfo, expected []string) bool {
	if len(expected) == 0 {
		return true
	}
	for _, source := range sources {
		preview := strings.ToLower(source.Preview)
		for _, term := range expected {
			if strings.Contains(preview, strings.ToLower(term)) {
				return true
			}
		}
	}
	return false
}
