package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type uploadedFile struct {
	Data        []byte
	Name        string
	ContentType string
	Metadata    map[string]interface{}
}

func (srv *APIServer) parseAndValidateFile(r *http.Request) (*uploadedFile, error) {
	if !strings.HasPrefix(r.Header.Get(contentTypeHeader), "multipart/form-data") {
		return nil, fmt.Errorf("content type must be multipart/form-data")
	}

	if err := r.ParseMultipartForm(srv.cfg.MaxUploadBytes); err != nil {
		return nil, fmt.Errorf("failed to parse form: %v", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("no file provided")
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	if len(fileData) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	metadata, err := parseCurationMetadata(r.FormValue("metadata"))
	if err != nil {
		return nil, err
	}

	return &uploadedFile{
		Data:        fileData,
		Name:        header.Filename,
		ContentType: header.Header.Get(contentTypeHeader),
		Metadata:    metadata,
	}, nil
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

	preferences, err := normalizeAskPreferences(req.Preferences)
	if err != nil {
		return nil, err
	}
	req.Preferences = preferences

	return &req, nil
}

func normalizeAskPreferences(preferences *AskPreferences) (*AskPreferences, error) {
	if preferences == nil {
		return nil, nil
	}

	var err error
	normalized := &AskPreferences{}
	if normalized.Layers, err = normalizePreferenceValues("layers", preferences.Layers, metadataEnums["layer"]); err != nil {
		return nil, err
	}
	if normalized.Categories, err = normalizePreferenceValues("categories", preferences.Categories, metadataEnums["category"]); err != nil {
		return nil, err
	}
	if normalized.Platforms, err = normalizePreferenceValues("platforms", preferences.Platforms, metadataEnums["platform"]); err != nil {
		return nil, err
	}
	if normalized.SourceKinds, err = normalizePreferenceValues("source_kinds", preferences.SourceKinds, metadataEnums["source_kind"]); err != nil {
		return nil, err
	}
	if normalized.SourceQuality, err = normalizePreferenceValues("source_quality", preferences.SourceQuality, metadataEnums["source_quality"]); err != nil {
		return nil, err
	}

	if len(normalized.Layers) == 0 &&
		len(normalized.Categories) == 0 &&
		len(normalized.Platforms) == 0 &&
		len(normalized.SourceKinds) == 0 &&
		len(normalized.SourceQuality) == 0 {
		return nil, nil
	}

	return normalized, nil
}

func normalizePreferenceValues(field string, values []string, allowed map[string]struct{}) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := allowed[value]; !ok {
			return nil, fmt.Errorf("preferences field %q has invalid value %q", field, value)
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, nil
}
