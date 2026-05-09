package loader

import (
	"context"
	"errors"
	"testing"
)

func TestMarkdownLoaderLoadsMarkdown(t *testing.T) {
	loaded, err := MarkdownLoader{}.Load(context.Background(), Source{
		Name:        "strategy.md",
		ContentType: "text/markdown",
		Data:        []byte("# Content Strategy\n\nWrite useful notes."),
	})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Document.Source != "markdown" {
		t.Fatalf("expected markdown source, got %s", loaded.Document.Source)
	}
	if loaded.Document.Title != "Content Strategy" {
		t.Fatalf("expected title from heading, got %s", loaded.Document.Title)
	}
}

func TestMarkdownLoaderSupportsMarkdownExtension(t *testing.T) {
	loader := MarkdownLoader{}

	if !loader.Supports(Source{Name: "readme.markdown"}) {
		t.Fatal("expected .markdown support")
	}
}

func TestMarkdownLoaderRejectsEmptyMarkdown(t *testing.T) {
	_, err := MarkdownLoader{}.Load(context.Background(), Source{
		Name: "empty.md",
		Data: []byte("  "),
	})
	if err == nil {
		t.Fatal("expected empty markdown error")
	}
	if !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("expected ErrInvalidDocument, got %v", err)
	}
}
