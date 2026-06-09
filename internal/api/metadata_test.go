package api

import "testing"

func TestParseCurationMetadataAcceptsValidMetadata(t *testing.T) {
	metadata, err := parseCurationMetadata(`{
		"type": "knowledge_asset",
		"layer": "foundations",
		"category": "storytelling",
		"platform": "general",
		"source_kind": "article",
		"source_quality": "high",
		"evergreen": true,
		"visibility": "private",
		"source_url": "https://example.com/article#section",
		"author": "Example Author",
		"created": "2026-05-27",
		"captured_at": "2026-05-27",
		"tags": ["narrative", "retention", "narrative"]
	}`)
	if err != nil {
		t.Fatalf("parseCurationMetadata failed: %v", err)
	}

	if metadata["layer"] != "foundations" {
		t.Fatalf("unexpected layer: %v", metadata["layer"])
	}
	if metadata["evergreen"] != true {
		t.Fatalf("unexpected evergreen: %v", metadata["evergreen"])
	}

	tags, ok := metadata["tags"].([]string)
	if !ok {
		t.Fatalf("expected tags to be []string, got %T", metadata["tags"])
	}
	if len(tags) != 2 {
		t.Fatalf("expected duplicate tags to be removed, got %v", tags)
	}
}

func TestParseCurationMetadataRejectsReservedFields(t *testing.T) {
	_, err := parseCurationMetadata(`{"filename":"fake.md"}`)
	if err == nil {
		t.Fatal("expected reserved field error")
	}
}

func TestParseCurationMetadataRejectsUnknownFields(t *testing.T) {
	_, err := parseCurationMetadata(`{"ingestion_status":"approved"}`)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestParseCurationMetadataRejectsInvalidEnum(t *testing.T) {
	_, err := parseCurationMetadata(`{"layer":"random"}`)
	if err == nil {
		t.Fatal("expected invalid enum error")
	}
}

func TestParseCurationMetadataRejectsInvalidTagType(t *testing.T) {
	_, err := parseCurationMetadata(`{"tags":["valid", 123]}`)
	if err == nil {
		t.Fatal("expected invalid tag type error")
	}
}

func TestParseCurationMetadataRejectsRelativeSourceURL(t *testing.T) {
	_, err := parseCurationMetadata(`{"source_url":"/article#section"}`)
	if err == nil {
		t.Fatal("expected relative URL error")
	}
}

func TestParseCurationMetadataRejectsUnsupportedSourceURLScheme(t *testing.T) {
	_, err := parseCurationMetadata(`{"source_url":"ftp://example.com/article"}`)
	if err == nil {
		t.Fatal("expected unsupported URL scheme error")
	}
}

func TestParseCurationMetadataReturnsNilForEmptyInput(t *testing.T) {
	metadata, err := parseCurationMetadata(" ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metadata != nil {
		t.Fatalf("expected nil metadata, got %v", metadata)
	}
}
