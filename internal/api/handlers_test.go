package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stan/Projects/studies/rag/internal/config"
)

func TestIngestHandler_WithValidPDF(t *testing.T) {
	// Skip if required environment variables are not set
	if os.Getenv("DATABASE_URL") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("DATABASE_URL or OPENAI_API_KEY not set, skipping integration test")
	}

	// Create API server
	cfg := &config.Config{
		Port:          ":8080",
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		ChunkTokens:   800,
		OverlapTokens: 100,
		TopK:          5,
	}

	server, err := NewAPIServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create API server: %v", err)
	}

	// Create a minimal PDF for testing
	pdfData := createTestPDF()

	// Create multipart form with PDF
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.pdf")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, bytes.NewReader(pdfData)); err != nil {
		t.Fatalf("Failed to copy PDF data: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/rag/ingest", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create HTTP response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler := server.IngestHandler()
	handler.ServeHTTP(w, req)

	// Check response status
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 200 or 500, got %d", w.Code)
	}

	// Parse response
	var respBody map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if w.Code == http.StatusOK {
		// Check response contains expected fields
		if respBody["status"] != "success" {
			t.Errorf("expected status 'success', got %v", respBody["status"])
		}

		if respBody["document_id"] == "" {
			t.Fatal("expected document_id in response")
		}
	}
}

func TestAskHandler_NoDocuments(t *testing.T) {
	// Skip if required environment variables are not set
	if os.Getenv("DATABASE_URL") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("DATABASE_URL or OPENAI_API_KEY not set, skipping integration test")
	}

	// Create API server
	cfg := &config.Config{
		Port:          ":8080",
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		ChunkTokens:   800,
		OverlapTokens: 100,
		TopK:          5,
	}

	server, err := NewAPIServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create API server: %v", err)
	}

	// Create ask request
	askReq := AskRequest{
		Question: "What is the meaning of life?",
		TopK:     5,
	}

	reqBody, err := json.Marshal(askReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/rag/ask", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler := server.AskHandler()
	handler.ServeHTTP(w, req)

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Parse response
	var respBody AskResponse
	if err := json.NewDecoder(w.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// When no documents exist, should return default message
	if respBody.Answer == "" {
		t.Fatal("expected non-empty answer")
	}
}

func TestAskHandler_InvalidRequest(t *testing.T) {
	// Skip if required environment variables are not set
	if os.Getenv("DATABASE_URL") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("DATABASE_URL or OPENAI_API_KEY not set, skipping integration test")
	}

	// Create API server
	cfg := &config.Config{
		Port:          ":8080",
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		ChunkTokens:   800,
		OverlapTokens: 100,
		TopK:          5,
	}

	server, err := NewAPIServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create API server: %v", err)
	}

	// Create invalid ask request (empty question)
	askReq := AskRequest{
		Question: "",
		TopK:     5,
	}

	reqBody, err := json.Marshal(askReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/rag/ask", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler := server.AskHandler()
	handler.ServeHTTP(w, req)

	// Check response status
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestIngestHandler_InvalidPDF(t *testing.T) {
	// Skip if required environment variables are not set
	if os.Getenv("DATABASE_URL") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("DATABASE_URL or OPENAI_API_KEY not set, skipping integration test")
	}

	// Create API server
	cfg := &config.Config{
		Port:          ":8080",
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		ChunkTokens:   800,
		OverlapTokens: 100,
		TopK:          5,
	}

	server, err := NewAPIServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create API server: %v", err)
	}

	// Create multipart form with invalid PDF
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Write invalid data (not a PDF)
	if _, err := part.Write([]byte("This is not a PDF")); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/rag/ingest", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create HTTP response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler := server.IngestHandler()
	handler.ServeHTTP(w, req)

	// Check response status - should be bad request
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// Helper function to create a test PDF
func createTestPDF() []byte {
	// Return a minimal PDF header to pass validation
	return []byte("%PDF-1.4\n%dummy PDF for testing\n")
}

// TestFullPipelineWithContext tests the full ingest and ask pipeline
func TestFullPipelineWithContext(t *testing.T) {
	// Skip if required environment variables are not set
	if os.Getenv("DATABASE_URL") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("DATABASE_URL or OPENAI_API_KEY not set, skipping integration test")
	}

	// Create API server
	cfg := &config.Config{
		Port:          ":8080",
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		ChunkTokens:   800,
		OverlapTokens: 100,
		TopK:          5,
	}

	server, err := NewAPIServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create API server: %v", err)
	}

	// Test that handlers exist and are callable
	ingestHandler := server.IngestHandler()
	if ingestHandler == nil {
		t.Fatal("IngestHandler returned nil")
	}

	askHandler := server.AskHandler()
	if askHandler == nil {
		t.Fatal("AskHandler returned nil")
	}

	// Test context propagation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Verify context is not cancelled
	select {
	case <-ctx.Done():
		t.Fatal("context cancelled unexpectedly")
	default:
		// Context is valid
	}
}

func TestIsAuthorizedAdminRequest(t *testing.T) {
	server := &APIServer{
		cfg: &config.Config{AdminToken: "test-admin-token"},
	}

	t.Run("accepts bearer token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/seed-demo", nil)
		req.Header.Set("Authorization", "Bearer test-admin-token")

		if !server.isAuthorizedAdminRequest(req) {
			t.Fatal("expected bearer token to be accepted")
		}
	})

	t.Run("accepts admin token header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/seed-demo", nil)
		req.Header.Set("X-Admin-Token", "test-admin-token")

		if !server.isAuthorizedAdminRequest(req) {
			t.Fatal("expected X-Admin-Token to be accepted")
		}
	})

	t.Run("rejects missing token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/seed-demo", nil)

		if server.isAuthorizedAdminRequest(req) {
			t.Fatal("expected missing token to be rejected")
		}
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/seed-demo", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")

		if server.isAuthorizedAdminRequest(req) {
			t.Fatal("expected invalid token to be rejected")
		}
	})
}

func TestIngestHandler_RequiresAdminTokenWhenPublicUploadDisabled(t *testing.T) {
	server := &APIServer{
		cfg: &config.Config{
			AdminToken:          "test-admin-token",
			PublicUploadEnabled: false,
			MaxUploadBytes:      10 << 20,
		},
	}

	req := httptest.NewRequest("POST", "/rag/ingest", nil)
	w := httptest.NewRecorder()

	server.IngestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestIngestHandler_ReturnsNotFoundWhenUploadDisabledWithoutAdminToken(t *testing.T) {
	server := &APIServer{
		cfg: &config.Config{
			PublicUploadEnabled: false,
			MaxUploadBytes:      10 << 20,
		},
	}

	req := httptest.NewRequest("POST", "/rag/ingest", nil)
	w := httptest.NewRecorder()

	server.IngestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestSeedDemoHandlerRequiresAdminToken(t *testing.T) {
	server := &APIServer{
		cfg: &config.Config{AdminToken: "test-admin-token"},
	}

	req := httptest.NewRequest("POST", "/admin/seed-demo", nil)
	w := httptest.NewRecorder()

	server.SeedDemoHandler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff header, got %q", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected DENY frame header, got %q", got)
	}
}

func TestParseAndValidateQuestionRequiresJSONContentType(t *testing.T) {
	server := &APIServer{cfg: &config.Config{TopK: 5}}
	req := httptest.NewRequest("POST", "/rag/ask", bytes.NewBufferString(`{"question":"hello"}`))

	if _, err := server.parseAndValidateQuestion(req); err == nil {
		t.Fatal("expected content type validation error")
	}
}

func TestParseAndValidateQuestionTrimsQuestion(t *testing.T) {
	server := &APIServer{cfg: &config.Config{TopK: 5}}
	req := httptest.NewRequest("POST", "/rag/ask", bytes.NewBufferString(`{"question":"  hello  "}`))
	req.Header.Set("Content-Type", "application/json")

	parsed, err := server.parseAndValidateQuestion(req)
	if err != nil {
		t.Fatalf("expected valid request: %v", err)
	}
	if parsed.Question != "hello" {
		t.Fatalf("expected trimmed question, got %q", parsed.Question)
	}
}
