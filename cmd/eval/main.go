package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	rageval "github.com/stan/Projects/studies/rag/internal/eval"
)

func main() {
	apiURL := flag.String("api-url", "http://localhost:8080", "RAG API base URL")
	questionsPath := flag.String("questions", "eval/questions.yaml", "evaluation questions YAML")
	githubSummary := flag.String("github-summary", os.Getenv("GITHUB_STEP_SUMMARY"), "GitHub step summary path")
	failOnThreshold := flag.Bool("fail-on-threshold", true, "exit non-zero when thresholds fail")
	flag.Parse()

	dataset, err := rageval.LoadDataset(*questionsPath)
	if err != nil {
		fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	report, err := rageval.NewRunner(*apiURL).Run(ctx, dataset)
	if err != nil {
		fatal(err)
	}

	fmt.Println(report.Markdown())
	if err := rageval.WriteGitHubSummary(*githubSummary, report); err != nil {
		fatal(err)
	}
	if *failOnThreshold && !report.Passed {
		os.Exit(1)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
