package loader

import (
	"context"
	"errors"
	"testing"
)

func TestTextLoaderLoadsPlainText(t *testing.T) {
	loaded, err := TextLoader{}.Load(context.Background(), Source{
		Name:        "notes.txt",
		ContentType: "text/plain",
		Data:        []byte("hello from text"),
	})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Document.Source != "text" {
		t.Fatalf("expected text source, got %s", loaded.Document.Source)
	}
	if loaded.Text != "hello from text" {
		t.Fatalf("unexpected text: %q", loaded.Text)
	}
}

func TestTextLoaderRejectsEmptyText(t *testing.T) {
	_, err := TextLoader{}.Load(context.Background(), Source{
		Name: "empty.txt",
		Data: []byte("  "),
	})
	if err == nil {
		t.Fatal("expected empty text error")
	}
	if !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("expected ErrInvalidDocument, got %v", err)
	}
}
