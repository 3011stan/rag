package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (srv *APIServer) parseAndValidateFile(r *http.Request) ([]byte, string, error) {
	if !strings.HasPrefix(r.Header.Get(contentTypeHeader), "multipart/form-data") {
		return nil, "", fmt.Errorf("content type must be multipart/form-data")
	}

	if err := r.ParseMultipartForm(srv.cfg.MaxUploadBytes); err != nil {
		return nil, "", fmt.Errorf("failed to parse form: %v", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, "", fmt.Errorf("no file provided")
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %v", err)
	}
	if len(fileData) == 0 {
		return nil, "", fmt.Errorf("file is empty")
	}
	if len(fileData) < 4 || string(fileData[:4]) != "%PDF" {
		return nil, "", fmt.Errorf("file is not a valid PDF")
	}

	return fileData, header.Filename, nil
}

func (srv *APIServer) parseAndValidateQuestion(r *http.Request) (*AskRequest, error) {
	if !strings.HasPrefix(r.Header.Get(contentTypeHeader), contentTypeJSON) {
		return nil, fmt.Errorf("content type must be application/json")
	}

	var req AskRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid request format: %v", err)
	}

	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		return nil, fmt.Errorf("question is required")
	}
	if len(req.Question) > maxQuestionLength {
		return nil, fmt.Errorf("question is too long (max %d chars)", maxQuestionLength)
	}

	if req.TopK <= 0 {
		req.TopK = srv.cfg.TopK
	}
	if req.TopK > maxTopK {
		req.TopK = maxTopK
	}

	return &req, nil
}
