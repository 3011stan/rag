package qa

import (
	"bytes"
	"text/template"
)

// PromptTemplate define um template de prompt para o LLM
type PromptTemplate struct {
	name     string
	template *template.Template
}

// PromptData contém dados para renderizar um prompt
type PromptData struct {
	Question string
	Context  string
	Sources  []SourceInfo
}

// SourceInfo contém informações sobre uma fonte
type SourceInfo struct {
	DocumentID string
	ChunkIndex int
	Score      float64
}

// NewPromptTemplate cria um novo template de prompt
func NewPromptTemplate(name string, tmpl string) (*PromptTemplate, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return nil, err
	}

	return &PromptTemplate{
		name:     name,
		template: t,
	}, nil
}

// Render renderiza o template com os dados fornecidos
func (pt *PromptTemplate) Render(data PromptData) (string, error) {
	var buf bytes.Buffer
	err := pt.template.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// DefaultPromptTemplate retorna o template padrão para RAG
func DefaultPromptTemplate() *PromptTemplate {
	tmpl := `You are a helpful AI assistant. Answer the user's question based on the provided context.

Question: {{.Question}}

Context:
{{.Context}}

Instructions:
- Provide a clear, concise answer based on the context above.
- If the context doesn't contain enough information to answer the question, say so.
- Be factual and avoid speculation.
- Format your answer clearly with paragraphs.

Answer:`

	pt, err := NewPromptTemplate("default_rag", tmpl)
	if err != nil {
		panic(err)
	}
	return pt
}

// DetailedPromptTemplate retorna um template mais detalhado com sources
func DetailedPromptTemplate() *PromptTemplate {
	tmpl := `You are a knowledgeable AI assistant. Answer the user's question based on the provided context and sources.

Question: {{.Question}}

Context from retrieved documents:
{{.Context}}

Source Information:
{{range .Sources}}- Document: {{.DocumentID}}, Chunk: {{.ChunkIndex}}, Relevance Score: {{printf "%.2f" .Score}}
{{end}}

Instructions:
- Provide a comprehensive answer based on the context above.
- Cite specific sources when relevant (e.g., "According to [DocumentID]...").
- If the context doesn't contain enough information, acknowledge this and provide any relevant general knowledge.
- Be accurate and avoid making up information.
- Structure your answer clearly with multiple paragraphs if needed.

Answer:`

	pt, err := NewPromptTemplate("detailed_rag", tmpl)
	if err != nil {
		panic(err)
	}
	return pt
}

// CustomPromptTemplate retorna um template customizável
func CustomPromptTemplate(template string) (*PromptTemplate, error) {
	return NewPromptTemplate("custom", template)
}

// SystemPrompt retorna o system prompt para o LLM
func SystemPrompt() string {
	return `You are an expert AI assistant specialized in analyzing documents and answering questions based on provided context. 
Your responses should be:
- Accurate and factual
- Clear and well-structured
- Based primarily on the provided context
- Helpful and informative

If you cannot answer based on the provided context, say so clearly.`
}

// BuildPrompt constrói um prompt completo para o LLM
func BuildPrompt(template *PromptTemplate, question string, context string, sources []SourceInfo) (string, error) {
	if template == nil {
		template = DefaultPromptTemplate()
	}

	data := PromptData{
		Question: question,
		Context:  context,
		Sources:  sources,
	}

	return template.Render(data)
}

// FormatContext formata um contexto a partir de chunks
func FormatContext(chunks []string) string {
	var buf bytes.Buffer
	for i, chunk := range chunks {
		if i > 0 {
			buf.WriteString("\n\n---\n\n")
		}
		buf.WriteString(chunk)
	}
	return buf.String()
}
