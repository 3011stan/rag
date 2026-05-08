package loader

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNewPDFLoader(t *testing.T) {
	loader := NewPDFLoader(0)
	if loader == nil {
		t.Fatal("NewPDFLoader returned nil")
	}
}

func TestLoadPDF_EmptyData(t *testing.T) {
	loader := NewPDFLoader(0)

	doc, text, err := loader.LoadPDF([]byte{}, "test.pdf")

	if err == nil {
		t.Fatal("expected error for empty PDF data")
	}

	if doc != nil {
		t.Fatal("expected nil document for empty data")
	}

	if text != "" {
		t.Fatal("expected empty text for empty data")
	}
}

func TestLoadPDF_InvalidPDF(t *testing.T) {
	loader := NewPDFLoader(0)

	// Create data that's not a valid PDF
	invalidData := []byte("This is not a PDF file")

	doc, text, err := loader.LoadPDF(invalidData, "test.txt")

	if err == nil {
		t.Fatal("expected error for invalid PDF data")
	}

	if doc != nil {
		t.Fatal("expected nil document for invalid data")
	}

	if text != "" {
		t.Fatal("expected empty text for invalid data")
	}
}

// Helper to create a minimal valid PDF
func createMinimalPDF() []byte {
	pdf := bytes.NewBuffer(nil)
	pdf.WriteString("%PDF-1.4\n")

	var offsets []int
	writeObj := func(body string) {
		offsets = append(offsets, pdf.Len())
		pdf.WriteString(body)
	}

	content := "BT\n/F1 12 Tf\n100 700 Td\n(Hello, World!) Tj\nET\n"
	writeObj("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	writeObj("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	writeObj("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n")
	writeObj(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(content), content))
	writeObj("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	startXRef := pdf.Len()
	pdf.WriteString("xref\n")
	pdf.WriteString("0 6\n")
	pdf.WriteString("0000000000 65535 f\n")
	for _, offset := range offsets {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n\n", offset))
	}
	pdf.WriteString("trailer\n")
	pdf.WriteString("<< /Size 6 /Root 1 0 R >>\n")
	pdf.WriteString("startxref\n")
	pdf.WriteString(fmt.Sprintf("%d\n", startXRef))
	pdf.WriteString("%%EOF\n")

	return pdf.Bytes()
}

func TestLoadPDF_ValidPDF(t *testing.T) {
	loader := NewPDFLoader(0)

	pdfData := createMinimalPDF()

	doc, text, err := loader.LoadPDF(pdfData, "test.pdf")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc == nil {
		t.Fatal("expected document to be non-nil")
	}

	// Check document ID is generated
	if doc.ID == "" {
		t.Fatal("expected document ID to be generated")
	}

	// Check source is set
	if doc.Source != "pdf" {
		t.Errorf("expected source 'pdf', got '%s'", doc.Source)
	}

	if text == "" {
		t.Fatal("expected extracted text")
	}
}

func TestLoadPDFFromFile_NonExistent(t *testing.T) {
	loader := NewPDFLoader(0)

	doc, text, err := loader.LoadPDFFromFile("/non/existent/file.pdf")

	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if doc != nil {
		t.Fatal("expected nil document for non-existent file")
	}

	if text != "" {
		t.Fatal("expected empty text for non-existent file")
	}
}
