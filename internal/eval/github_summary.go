package eval

import (
	"fmt"
	"os"
	"strings"
)

func (r Report) Markdown() string {
	status := "PASS"
	if !r.Passed {
		status = "FAIL"
	}

	var b strings.Builder
	b.WriteString("# RAG Retrieval Eval\n\n")
	fmt.Fprintf(&b, "**Status:** %s\n\n", status)
	fmt.Fprintf(&b, "- Questions: %d\n", r.QuestionsTotal)
	fmt.Fprintf(&b, "- Metadata hit rate: %.2f\n", r.MetadataHitRate)
	fmt.Fprintf(&b, "- Top source metadata hit rate: %.2f\n", r.TopSourceMetadataHitRate)
	fmt.Fprintf(&b, "- Empty source count: %d\n\n", r.EmptySourceCount)

	b.WriteString("## Thresholds\n\n")
	fmt.Fprintf(&b, "- Metadata hit rate >= %.2f\n", r.Thresholds.MetadataHitRate)
	fmt.Fprintf(&b, "- Top source metadata hit rate >= %.2f\n", r.Thresholds.TopSourceMetadataHitRate)
	fmt.Fprintf(&b, "- Min sources per question >= %d\n\n", r.Thresholds.MinSourcesPerQuestion)

	b.WriteString("## Results\n\n")
	b.WriteString("| Question | Status | Sources | Metadata Hit | Top Source Hit | Preview Hit | Missing Metadata |\n")
	b.WriteString("| --- | --- | ---: | --- | --- | --- | --- |\n")
	for _, result := range r.Results {
		fmt.Fprintf(
			&b,
			"| %s | %s | %d | %s | %s | %s | %s |\n",
			result.ID,
			result.Status,
			result.SourceCount,
			boolText(result.MetadataHit),
			boolText(result.TopSourceMetadataHit),
			boolText(result.PreviewHit),
			strings.Join(result.MissingExpectedMetadata, ", "),
		)
	}
	return b.String()
}

func WriteGitHubSummary(path string, report *Report) error {
	if path == "" {
		return nil
	}
	return os.WriteFile(path, []byte(report.Markdown()), 0o644)
}

func boolText(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
