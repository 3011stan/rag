package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stan/Projects/studies/rag/internal/config"
)

func TestParseAndValidateFileAcceptsOptionalMetadata(t *testing.T) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("metadata", `{"layer":"foundations","category":"storytelling","tags":["narrative"]}`); err != nil {
		t.Fatalf("failed to write metadata: %v", err)
	}
	part, err := writer.CreateFormFile("file", "notes.md")
	if err != nil {
		t.Fatalf("failed to create file part: %v", err)
	}
	if _, err := part.Write([]byte("# Notes")); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/rag/ingest", body)
	req.Header.Set(contentTypeHeader, writer.FormDataContentType())

	srv := &APIServer{cfg: &config.Config{MaxUploadBytes: 1024}}
	uploaded, err := srv.parseAndValidateFile(req)
	if err != nil {
		t.Fatalf("parseAndValidateFile failed: %v", err)
	}

	if uploaded.Name != "notes.md" {
		t.Fatalf("unexpected filename: %s", uploaded.Name)
	}
	if uploaded.Metadata["layer"] != "foundations" {
		t.Fatalf("expected metadata layer, got %v", uploaded.Metadata["layer"])
	}
}

func TestParseAndValidateFileRejectsInvalidMetadata(t *testing.T) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("metadata", `{"source_type":"article"}`); err != nil {
		t.Fatalf("failed to write metadata: %v", err)
	}
	part, err := writer.CreateFormFile("file", "notes.md")
	if err != nil {
		t.Fatalf("failed to create file part: %v", err)
	}
	if _, err := part.Write([]byte("# Notes")); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/rag/ingest", body)
	req.Header.Set(contentTypeHeader, writer.FormDataContentType())

	srv := &APIServer{cfg: &config.Config{MaxUploadBytes: 1024}}
	if _, err := srv.parseAndValidateFile(req); err == nil {
		t.Fatal("expected invalid metadata error")
	}
}

func TestParseAndValidateQuestionAcceptsPreferences(t *testing.T) {
	srv := &APIServer{cfg: &config.Config{TopK: 5}}
	req := httptest.NewRequest(http.MethodPost, "/rag/ask", bytes.NewBufferString(`{
		"question":"How should I structure a technical reel?",
		"top_k":3,
		"preferences":{
			"layers":["platform_specific","self_knowledge","platform_specific"],
			"categories":["storytelling"],
			"platforms":["reels"],
			"source_kinds":["note","article"],
			"source_quality":["high"]
		}
	}`))
	req.Header.Set(contentTypeHeader, contentTypeJSON)

	parsed, err := srv.parseAndValidateQuestion(req)
	if err != nil {
		t.Fatalf("expected valid preferences: %v", err)
	}
	if parsed.Preferences == nil {
		t.Fatal("expected preferences")
	}
	if len(parsed.Preferences.Layers) != 2 {
		t.Fatalf("expected duplicate layers to be removed, got %v", parsed.Preferences.Layers)
	}
	if parsed.Preferences.Platforms[0] != "reels" {
		t.Fatalf("expected reels platform, got %v", parsed.Preferences.Platforms)
	}
	if parsed.Preferences.SourceKinds[1] != "article" {
		t.Fatalf("expected source kinds to be preserved, got %v", parsed.Preferences.SourceKinds)
	}
}

func TestParseAndValidateQuestionNormalizesEmptyPreferences(t *testing.T) {
	srv := &APIServer{cfg: &config.Config{TopK: 5}}
	req := httptest.NewRequest(http.MethodPost, "/rag/ask", bytes.NewBufferString(`{
		"question":"hello",
		"preferences":{
			"layers":[" ",""],
			"categories":[],
			"platforms":[]
		}
	}`))
	req.Header.Set(contentTypeHeader, contentTypeJSON)

	parsed, err := srv.parseAndValidateQuestion(req)
	if err != nil {
		t.Fatalf("expected empty preferences to normalize: %v", err)
	}
	if parsed.Preferences != nil {
		t.Fatalf("expected nil preferences after normalization, got %v", parsed.Preferences)
	}
}

func TestParseAndValidateQuestionRejectsInvalidPreferences(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid layer",
			body: `{"question":"hello","preferences":{"layers":["invalid"]}}`,
		},
		{
			name: "invalid category",
			body: `{"question":"hello","preferences":{"categories":["invalid"]}}`,
		},
		{
			name: "invalid platform",
			body: `{"question":"hello","preferences":{"platforms":["invalid"]}}`,
		},
		{
			name: "invalid source kind",
			body: `{"question":"hello","preferences":{"source_kinds":["invalid"]}}`,
		},
		{
			name: "invalid source quality",
			body: `{"question":"hello","preferences":{"source_quality":["invalid"]}}`,
		},
		{
			name: "unknown preference field",
			body: `{"question":"hello","preferences":{"visibility":["public"]}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &APIServer{cfg: &config.Config{TopK: 5}}
			req := httptest.NewRequest(http.MethodPost, "/rag/ask", bytes.NewBufferString(tt.body))
			req.Header.Set(contentTypeHeader, contentTypeJSON)

			if _, err := srv.parseAndValidateQuestion(req); err == nil {
				t.Fatal("expected invalid preferences error")
			}
		})
	}
}
