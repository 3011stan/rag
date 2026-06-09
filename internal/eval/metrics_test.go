package eval

import "testing"

func TestEvaluateQuestionMatchesExpectedMetadata(t *testing.T) {
	question := Question{
		ID: "q001",
		ExpectedMetadata: ExpectedMetadata{
			Layers:    []string{"foundations"},
			Platforms: []string{"general"},
		},
		ExpectedPreviewContains: []string{"technical"},
	}
	response := &askResponse{
		Sources: []SourceInfo{
			{
				Preview: "technical concept explanation",
				Metadata: map[string]interface{}{
					"layer":    "foundations",
					"platform": "general",
				},
			},
		},
	}

	result := evaluateQuestion(question, response)

	if !result.MetadataHit {
		t.Fatal("expected metadata hit")
	}
	if !result.TopSourceMetadataHit {
		t.Fatal("expected top source metadata hit")
	}
	if !result.PreviewHit {
		t.Fatal("expected preview hit")
	}
	if result.Status != "pass" {
		t.Fatalf("expected pass status, got %q", result.Status)
	}
}

func TestSummarizeFailsBelowThreshold(t *testing.T) {
	report := summarize(Thresholds{
		MetadataHitRate:          1,
		TopSourceMetadataHitRate: 1,
		MinSourcesPerQuestion:    1,
	}, []QuestionResult{
		{ID: "q001", SourceCount: 1, MetadataHit: true, TopSourceMetadataHit: true},
		{ID: "q002", SourceCount: 1, MetadataHit: false, TopSourceMetadataHit: false},
	})

	if report.Passed {
		t.Fatal("expected report to fail thresholds")
	}
	if report.MetadataHitRate != 0.5 {
		t.Fatalf("expected metadata hit rate 0.5, got %v", report.MetadataHitRate)
	}
}
