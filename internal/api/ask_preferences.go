package api

import ragpipeline "github.com/stan/Projects/studies/rag/internal/rag/pipeline"

func toPipelinePreferences(preferences *AskPreferences) *ragpipeline.AskPreferences {
	if preferences == nil {
		return nil
	}

	return &ragpipeline.AskPreferences{
		Layers:        preferences.Layers,
		Categories:    preferences.Categories,
		Platforms:     preferences.Platforms,
		SourceKinds:   preferences.SourceKinds,
		SourceQuality: preferences.SourceQuality,
	}
}
