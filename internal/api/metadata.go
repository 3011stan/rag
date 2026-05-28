package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

var reservedMetadataFields = map[string]struct{}{
	"filename":     {},
	"content_type": {},
	"source_type":  {},
	"pages":        {},
	"checksum":     {},
}

var allowedMetadataFields = map[string]struct{}{
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
}

var metadataEnums = map[string]map[string]struct{}{
	"type": {
		"knowledge_asset":    {},
		"editorial_decision": {},
		"script":             {},
		"workflow":           {},
		"reference":          {},
	},
	"layer": {
		"foundations":         {},
		"platform_specific":   {},
		"self_knowledge":      {},
		"editorial_decisions": {},
	},
	"category": {
		"creator_systems":         {},
		"storytelling":            {},
		"technical_communication": {},
		"platform_dynamics":       {},
		"self_knowledge":          {},
		"editorial_decisions":     {},
	},
	"platform": {
		"general":  {},
		"youtube":  {},
		"reels":    {},
		"tiktok":   {},
		"linkedin": {},
		"medium":   {},
		"podcast":  {},
	},
	"source_kind": {
		"article":     {},
		"transcript":  {},
		"note":        {},
		"script":      {},
		"decision":    {},
		"pdf_extract": {},
		"workflow":    {},
		"reference":   {},
	},
	"source_quality": {
		"high":   {},
		"medium": {},
		"low":    {},
	},
	"visibility": {
		"private":        {},
		"portfolio_demo": {},
		"public":         {},
	},
}

func parseCurationMetadata(raw string) (map[string]interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return nil, fmt.Errorf("metadata must be valid JSON: %w", err)
	}
	if len(metadata) == 0 {
		return nil, nil
	}

	clean := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		if _, reserved := reservedMetadataFields[key]; reserved {
			return nil, fmt.Errorf("metadata field %q is reserved", key)
		}
		if _, allowed := allowedMetadataFields[key]; !allowed {
			return nil, fmt.Errorf("metadata field %q is not allowed", key)
		}

		normalized, err := normalizeMetadataValue(key, value)
		if err != nil {
			return nil, err
		}
		if normalized != nil {
			clean[key] = normalized
		}
	}

	if len(clean) == 0 {
		return nil, nil
	}
	return clean, nil
}

func normalizeMetadataValue(key string, value interface{}) (interface{}, error) {
	switch key {
	case "type", "layer", "category", "platform", "source_kind", "source_quality", "visibility":
		text, err := requiredMetadataString(key, value)
		if err != nil {
			return nil, err
		}
		if _, ok := metadataEnums[key][text]; !ok {
			return nil, fmt.Errorf("metadata field %q has invalid value %q", key, text)
		}
		return text, nil
	case "evergreen":
		boolean, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("metadata field %q must be a boolean", key)
		}
		return boolean, nil
	case "source_url":
		text, err := optionalMetadataString(key, value)
		if err != nil || text == "" {
			return text, err
		}
		parsed, parseErr := url.ParseRequestURI(text)
		if parseErr != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, fmt.Errorf("metadata field %q must be a valid absolute URL", key)
		}
		return text, nil
	case "author", "created", "captured_at":
		return optionalMetadataString(key, value)
	case "tags":
		return normalizeMetadataTags(value)
	default:
		return nil, fmt.Errorf("metadata field %q is not supported", key)
	}
}

func requiredMetadataString(key string, value interface{}) (string, error) {
	text, err := optionalMetadataString(key, value)
	if err != nil {
		return "", err
	}
	if text == "" {
		return "", fmt.Errorf("metadata field %q is required when present", key)
	}
	return text, nil
}

func optionalMetadataString(key string, value interface{}) (string, error) {
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("metadata field %q must be a string", key)
	}
	return strings.TrimSpace(text), nil
}

func normalizeMetadataTags(value interface{}) ([]string, error) {
	values, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("metadata field %q must be an array of strings", "tags")
	}

	tags := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		tag, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("metadata field %q must contain only strings", "tags")
		}
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}

	if len(tags) == 0 {
		return nil, nil
	}
	return tags, nil
}
