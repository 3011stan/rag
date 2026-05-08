package loader

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
	"github.com/stan/Projects/studies/rag/internal/rag"
)

// PDFMetadata contém metadados extraídos do PDF
type PDFMetadata struct {
	Filename string
	Pages    int
	Author   string
	Title    string
	Subject  string
	Keywords string
	Checksum string
}

// PDFLoader é responsável por carregar e extrair texto de PDFs
type PDFLoader struct {
	maxPages int // 0 = sem limite
}

// NewPDFLoader cria uma nova instância do PDFLoader
func NewPDFLoader(maxPages int) *PDFLoader {
	return &PDFLoader{
		maxPages: maxPages,
	}
}

// LoadPDF carrega um PDF de um buffer e extrai texto + metadados
func (pl *PDFLoader) LoadPDF(data []byte, filename string) (*rag.Document, string, error) {
	if len(data) == 0 {
		return nil, "", fmt.Errorf("empty PDF data")
	}

	// Extrair metadados
	metadata, err := pl.extractMetadata(data, filename)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Extrair texto
	text, err := pl.ExtractText(data, metadata.Pages)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract text: %w", err)
	}

	if strings.TrimSpace(text) == "" {
		return nil, "", fmt.Errorf("no text extracted from PDF")
	}

	// Criar documento
	doc := &rag.Document{
		ID:       generateID(filename, data),
		Source:   "pdf",
		Title:    metadata.Title,
		Checksum: metadata.Checksum,
		Metadata: map[string]interface{}{
			"filename": metadata.Filename,
			"pages":    metadata.Pages,
			"author":   metadata.Author,
			"subject":  metadata.Subject,
			"keywords": metadata.Keywords,
		},
	}

	return doc, text, nil
}

// extractMetadata extrai metadados do PDF
func (pl *PDFLoader) extractMetadata(data []byte, filename string) (*PDFMetadata, error) {
	// Calcular checksum
	checksum := fmt.Sprintf("%x", md5.Sum(data))

	// Ler arquivo PDF
	reader := bytes.NewReader(data)
	pdfFile, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	// Extrair metadados do PDF
	metadata := &PDFMetadata{
		Filename: filename,
		Pages:    pdfFile.NumPage(),
		Checksum: checksum,
	}

	// Tentar extrair informações do documento
	// Nota: A biblioteca ledongthuc/pdf tem suporte limitado a metadados
	// Metadados adicionais podem ser extraídos conforme necessário

	return metadata, nil
}

// ExtractText extrai todo o texto do PDF
func (pl *PDFLoader) ExtractText(data []byte, totalPages int) (string, error) {
	reader := bytes.NewReader(data)
	pdfFile, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to read PDF: %w", err)
	}

	// Determinar número de páginas a processar
	pagesToProcess := pdfFile.NumPage()
	if pl.maxPages > 0 && pagesToProcess > pl.maxPages {
		pagesToProcess = pl.maxPages
	}

	var textBuffer strings.Builder

	// Iterar sobre as páginas
	for pageNum := 1; pageNum <= pagesToProcess; pageNum++ {
		page := pdfFile.Page(pageNum)

		// Extrair texto da página
		text, err := page.GetPlainText(nil)
		if err != nil {
			// Log do erro mas continua processando
			fmt.Printf("warning: failed to extract text from page %d: %v\n", pageNum, err)
			continue
		}

		// Adicionar separador entre páginas
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

// generateID gera um ID único para o documento baseado no filename e conteúdo
func generateID(filename string, data []byte) string {
	idSource := append([]byte(filename), data...)
	return uuid.NewSHA1(uuid.NameSpaceURL, idSource).String()
}

// LoadPDFFromFile carrega um PDF de um arquivo do sistema
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
