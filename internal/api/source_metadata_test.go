package api

import "testing"

func TestSanitizeSourceMetadataAllowsPublicFields(t *testing.T) {
	metadata := map[string]interface{}{
		"type":        "knowledge_asset",
		"layer":       "foundations",
		"filename":    "notes.md",
		"source_type": "markdown",
		"pages":       3,
		"tags":        []string{"rag", "metadata"},
	}

	sanitized := sanitizeSourceMetadata(metadata)

	if sanitized["type"] != "knowledge_asset" {
		t.Fatalf("expected type to be preserved, got %v", sanitized["type"])
	}
	if sanitized["filename"] != "notes.md" {
		t.Fatalf("expected filename to be preserved, got %v", sanitized["filename"])
	}
	if sanitized["pages"] != 3 {
		t.Fatalf("expected pages to be preserved, got %v", sanitized["pages"])
	}
}

func TestSanitizeSourceMetadataRemovesPrivateFields(t *testing.T) {
	metadata := map[string]interface{}{
		"author":   "Stan",
		"checksum": "secret-checksum",
		"unknown":  "unexpected",
	}

	sanitized := sanitizeSourceMetadata(metadata)

	if sanitized["author"] != "Stan" {
		t.Fatalf("expected author to be preserved, got %v", sanitized["author"])
	}
	if _, ok := sanitized["checksum"]; ok {
		t.Fatal("expected checksum to be removed")
	}
	if _, ok := sanitized["unknown"]; ok {
		t.Fatal("expected unknown fields to be removed")
	}
}

func TestSanitizeSourceMetadataReturnsNilWhenEmpty(t *testing.T) {
	if sanitized := sanitizeSourceMetadata(nil); sanitized != nil {
		t.Fatalf("expected nil metadata, got %v", sanitized)
	}
	if sanitized := sanitizeSourceMetadata(map[string]interface{}{"checksum": "hidden"}); sanitized != nil {
		t.Fatalf("expected nil metadata with only private fields, got %v", sanitized)
	}
}
