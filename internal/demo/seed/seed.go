package seed

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/stan/Projects/studies/rag/internal/rag/loader"
	ragpipeline "github.com/stan/Projects/studies/rag/internal/rag/pipeline"
)

const DefaultDocsDir = "demo/docs"

func Directory(
	ctx context.Context,
	docsDir string,
	pipeline *ragpipeline.Pipeline,
) (int, error) {
	if docsDir == "" {
		docsDir = DefaultDocsDir
	}

	var seeded int
	err := filepath.WalkDir(docsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !isDemoDoc(path) {
			return nil
		}

		if err := file(ctx, path, pipeline); err != nil {
			return err
		}
		seeded++
		return nil
	})
	if err != nil {
		return 0, err
	}

	return seeded, nil
}

func file(
	ctx context.Context,
	path string,
	pipeline *ragpipeline.Pipeline,
) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	_, err = pipeline.Ingest(ctx, loader.Source{
		Name:        path,
		ContentType: contentType(path),
		Data:        data,
	})
	if err != nil {
		return fmt.Errorf("failed to ingest demo document %s: %w", path, err)
	}

	return nil
}

func isDemoDoc(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".txt":
		return true
	default:
		return false
	}
}

func contentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
