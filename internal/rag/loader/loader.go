package loader

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/stan/Projects/studies/rag/internal/rag"
)

var (
	ErrUnsupportedType = errors.New("unsupported document type")
	ErrInvalidDocument = errors.New("invalid document")
)

type Source struct {
	Name        string
	ContentType string
	Data        []byte
}

type LoadedDocument struct {
	Document rag.Document
	Text     string
}

type DocumentLoader interface {
	Supports(source Source) bool
	Load(ctx context.Context, source Source) (*LoadedDocument, error)
}

type Registry struct {
	loaders []DocumentLoader
}

func NewRegistry(loaders ...DocumentLoader) *Registry {
	return &Registry{loaders: loaders}
}

func DefaultRegistry() *Registry {
	return NewRegistry(
		NewPDFLoader(0),
		MarkdownLoader{},
		TextLoader{},
	)
}

func (r *Registry) Load(ctx context.Context, source Source) (*LoadedDocument, error) {
	for _, loader := range r.loaders {
		if loader.Supports(source) {
			return loader.Load(ctx, source)
		}
	}
	return nil, fmt.Errorf("%w for %s", ErrUnsupportedType, source.Name)
}

func Extension(source Source) string {
	return strings.ToLower(filepath.Ext(source.Name))
}

func BaseTitle(source Source) string {
	base := filepath.Base(source.Name)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
