package pipeline

import "testing"

func TestMergeMetadataKeepsTechnicalMetadata(t *testing.T) {
	base := map[string]interface{}{
		"filename":    "notes.md",
		"source_type": "markdown",
	}
	curation := map[string]interface{}{
		"filename": "manual.md",
		"layer":    "foundations",
	}

	merged := mergeMetadata(base, curation)

	if merged["filename"] != "notes.md" {
		t.Fatalf("expected technical filename to be preserved, got %v", merged["filename"])
	}
	if merged["source_type"] != "markdown" {
		t.Fatalf("expected source_type to be preserved, got %v", merged["source_type"])
	}
	if merged["layer"] != "foundations" {
		t.Fatalf("expected curation metadata to be merged, got %v", merged["layer"])
	}
}

func TestMergeMetadataHandlesEmptyMetadata(t *testing.T) {
	merged := mergeMetadata(nil, nil)
	if merged != nil {
		t.Fatalf("expected nil metadata, got %v", merged)
	}
}
