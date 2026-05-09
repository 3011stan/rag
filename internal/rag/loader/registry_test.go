package loader

import (
	"context"
	"errors"
	"testing"
)

func TestRegistrySelectsTextLoader(t *testing.T) {
	registry := NewRegistry(TextLoader{})

	loaded, err := registry.Load(context.Background(), Source{
		Name: "notes.txt",
		Data: []byte("hello"),
	})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Document.Source != "text" {
		t.Fatalf("expected text source, got %s", loaded.Document.Source)
	}
}

func TestRegistryRejectsUnsupportedType(t *testing.T) {
	registry := NewRegistry(TextLoader{})

	_, err := registry.Load(context.Background(), Source{
		Name: "data.json",
		Data: []byte(`{"hello":"world"}`),
	})
	if err == nil {
		t.Fatal("expected unsupported type error")
	}
	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}
