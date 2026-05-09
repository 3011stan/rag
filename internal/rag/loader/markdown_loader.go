package loader

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/stan/Projects/studies/rag/internal/rag"
)

type MarkdownLoader struct{}

func (MarkdownLoader) Supports(source Source) bool {
	contentType := strings.ToLower(source.ContentType)
	return Extension(source) == ".md" ||
		Extension(source) == ".markdown" ||
		strings.Contains(contentType, "markdown")
}

func (MarkdownLoader) Load(ctx context.Context, source Source) (*LoadedDocument, error) {
	_ = ctx

	text := strings.TrimSpace(string(source.Data))
	if text == "" {
		return nil, fmt.Errorf("%w: empty markdown document", ErrInvalidDocument)
	}

	checksum := fmt.Sprintf("%x", md5.Sum(source.Data))
	doc := rag.Document{
		ID:       uuid.NewSHA1(uuid.NameSpaceURL, []byte(source.Name+":"+checksum)).String(),
		Source:   "markdown",
		Title:    markdownTitle(text, BaseTitle(source)),
		Checksum: checksum,
		Metadata: map[string]interface{}{
			"filename":     source.Name,
			"content_type": source.ContentType,
			"source_type":  "markdown",
		},
	}

	return &LoadedDocument{Document: doc, Text: text}, nil
}

func markdownTitle(text, fallback string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			if title != "" {
				return title
			}
		}
	}
	return fallback
}
