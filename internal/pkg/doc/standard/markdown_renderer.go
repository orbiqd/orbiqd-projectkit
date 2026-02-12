package standard

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
	"time"

	standardAPI "github.com/orbiqd/orbiqd-projectkit/pkg/doc/standard"
)

//go:embed standard.go.tmpl
var templateContent string

type MarkdownRenderer struct {
}

var _ standardAPI.Renderer = (*MarkdownRenderer)(nil)

func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

var funcMap = template.FuncMap{
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	"join": func(slice []string, separator string) string {
		return strings.Join(slice, separator)
	},
	"add": func(a, b int) int {
		return a + b
	},
	"trimTrailingNewlines": func(s string) string {
		return strings.TrimRight(s, "\n")
	},
	"codeFence": func() string {
		return "```"
	},
	"now": func() string {
		return time.Now().Format("2006-01-02 15:04:05 MST")
	},
	"slugify": func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, " ", "-")
		s = strings.ReplaceAll(s, "*", "")
		s = strings.ReplaceAll(s, ":", "")
		s = strings.ReplaceAll(s, ",", "")
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, "(", "")
		s = strings.ReplaceAll(s, ")", "")
		s = strings.ReplaceAll(s, "'", "")
		s = strings.ReplaceAll(s, "\"", "")
		return s
	},
}

func (renderer *MarkdownRenderer) Render(standard standardAPI.Standard) ([]byte, error) {
	tmpl, err := template.New("standard.go.tmpl").Funcs(funcMap).Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("template parse: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, standard); err != nil {
		return nil, fmt.Errorf("template execution: %w", err)
	}

	return buf.Bytes(), nil
}

func (renderer *MarkdownRenderer) FileExtension() string {
	return ".md"
}
