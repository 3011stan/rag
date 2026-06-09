package pipeline

import "testing"

func TestMergeMetadataKeepsTechnicalMetadata(t *testing.T) {
	base := map[string]interface{}{
		"filename":    "notes.md",
		"source_type": "markdown",
		"author":      "Loader Author",
	}
	curation := map[string]interface{}{
		"filename": "manual.md",
		"layer":    "foundations",
		"author":   "Curated Author",
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
	if merged["author"] != "Curated Author" {
		t.Fatalf("expected curation author to override loader author, got %v", merged["author"])
	}
}

func TestMergeMetadataKeepsLoaderMetadataWhenCurationMissing(t *testing.T) {
	base := map[string]interface{}{
		"author": "Loader Author",
		"pages":  3,
	}

	merged := mergeMetadata(base, nil)

	if merged["author"] != "Loader Author" {
		t.Fatalf("expected loader author to be preserved, got %v", merged["author"])
	}
	if merged["pages"] != 3 {
		t.Fatalf("expected loader pages to be preserved, got %v", merged["pages"])
	}
}

func TestMergeMetadataHandlesEmptyMetadata(t *testing.T) {
	merged := mergeMetadata(nil, nil)
	if merged != nil {
		t.Fatalf("expected nil metadata, got %v", merged)
	}
}
