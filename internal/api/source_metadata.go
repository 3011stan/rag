package api

var publicSourceMetadataFields = map[string]struct{}{
	"type":           {},
	"layer":          {},
	"category":       {},
	"platform":       {},
	"source_kind":    {},
	"source_quality": {},
	"evergreen":      {},
	"visibility":     {},
	"source_url":     {},
	"author":         {},
	"created":        {},
	"captured_at":    {},
	"tags":           {},
	"source_type":    {},
	"content_type":   {},
	"filename":       {},
	"pages":          {},
}

func sanitizeSourceMetadata(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return nil
	}

	sanitized := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		if _, allowed := publicSourceMetadataFields[key]; allowed {
			sanitized[key] = value
		}
	}
	if len(sanitized) == 0 {
		return nil
	}
	return sanitized
}
