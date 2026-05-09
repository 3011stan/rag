package loader

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
	"github.com/stan/Projects/studies/rag/internal/rag"
)

type PDFMetadata struct {
	Filename string
	Pages    int
	Author   string
	Title    string
	Subject  string
	Keywords string
	Checksum string
}

type PDFLoader struct {
	maxPages int // 0 = sem limite
}

func NewPDFLoader(maxPages int) *PDFLoader {
	return &PDFLoader{
		maxPages: maxPages,
	}
}

func (pl *PDFLoader) Supports(source Source) bool {
	return Extension(source) == ".pdf" ||
		strings.HasPrefix(strings.ToLower(source.ContentType), "application/pdf") ||
		bytes.HasPrefix(source.Data, []byte("%PDF"))
}

func (pl *PDFLoader) Load(ctx context.Context, source Source) (*LoadedDocument, error) {
	_ = ctx

	doc, text, err := pl.LoadPDF(source.Data, source.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}

	return &LoadedDocument{
		Document: *doc,
		Text:     text,
	}, nil
}

func (pl *PDFLoader) LoadPDF(data []byte, filename string) (*rag.Document, string, error) {
	if len(data) == 0 {
		return nil, "", fmt.Errorf("empty PDF data")
	}

	metadata, err := pl.extractMetadata(data, filename)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract metadata: %w", err)
	}

	text, err := pl.ExtractText(data, metadata.Pages)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract text: %w", err)
	}

	if strings.TrimSpace(text) == "" {
		return nil, "", fmt.Errorf("no text extracted from PDF")
	}

	doc := &rag.Document{
		ID:       generateID(filename, data),
		Source:   "pdf",
		Title:    metadata.Title,
		Checksum: metadata.Checksum,
		Metadata: map[string]interface{}{
			"filename":    metadata.Filename,
			"pages":       metadata.Pages,
			"author":      metadata.Author,
			"subject":     metadata.Subject,
			"keywords":    metadata.Keywords,
			"source_type": "pdf",
		},
	}

	return doc, text, nil
}

func (pl *PDFLoader) extractMetadata(data []byte, filename string) (*PDFMetadata, error) {
	checksum := fmt.Sprintf("%x", md5.Sum(data))

	reader := bytes.NewReader(data)
	pdfFile, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	metadata := &PDFMetadata{
		Filename: filename,
		Pages:    pdfFile.NumPage(),
		Checksum: checksum,
	}

	return metadata, nil
}

func (pl *PDFLoader) ExtractText(data []byte, totalPages int) (string, error) {
	reader := bytes.NewReader(data)
	pdfFile, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to read PDF: %w", err)
	}

	pagesToProcess := pdfFile.NumPage()
	if pl.maxPages > 0 && pagesToProcess > pl.maxPages {
		pagesToProcess = pl.maxPages
	}

	var textBuffer strings.Builder

	for pageNum := 1; pageNum <= pagesToProcess; pageNum++ {
		page := pdfFile.Page(pageNum)

		text, err := page.GetPlainText(nil)
		if err != nil {
			fmt.Printf("warning: failed to extract text from page %d: %v\n", pageNum, err)
			continue
		}

		if pageNum > 1 {
			textBuffer.WriteString("\n\n--- PAGE ")
			textBuffer.WriteString(fmt.Sprintf("%d", pageNum))
			textBuffer.WriteString(" ---\n\n")
		} else {
			textBuffer.WriteString("--- PAGE 1 ---\n\n")
		}

		textBuffer.WriteString(text)
	}

	text := textBuffer.String()
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("no text could be extracted from PDF")
	}

	return text, nil
}

func generateID(filename string, data []byte) string {
	idSource := append([]byte(filename), data...)
	return uuid.NewSHA1(uuid.NameSpaceURL, idSource).String()
}

func (pl *PDFLoader) LoadPDFFromFile(filePath string) (*rag.Document, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	return pl.LoadPDF(data, filePath)
}
