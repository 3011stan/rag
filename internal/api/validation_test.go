package api

import (
	"bytes"
	"mime/multipart"
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
