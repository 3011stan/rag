package loader

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/stan/Projects/studies/rag/internal/rag"
)

type TextLoader struct{}

func (TextLoader) Supports(source Source) bool {
	return Extension(source) == ".txt" || strings.HasPrefix(source.ContentType, "text/plain")
}

func (TextLoader) Load(ctx context.Context, source Source) (*LoadedDocument, error) {
	_ = ctx

	text := strings.TrimSpace(string(source.Data))
	if text == "" {
		return nil, fmt.Errorf("%w: empty text document", ErrInvalidDocument)
	}

	checksum := fmt.Sprintf("%x", md5.Sum(source.Data))
	doc := rag.Document{
		ID:       uuid.NewSHA1(uuid.NameSpaceURL, []byte(source.Name+":"+checksum)).String(),
		Source:   "text",
		Title:    BaseTitle(source),
		Checksum: checksum,
		Metadata: map[string]interface{}{
			"filename":     source.Name,
			"content_type": source.ContentType,
			"source_type":  "text",
		},
	}

	return &LoadedDocument{Document: doc, Text: text}, nil
}
