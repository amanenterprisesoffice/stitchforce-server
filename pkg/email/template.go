package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"stitchforce-server/pkg/globals"
	"time"
)

//go:embed templates/email.html
var templateFiles embed.FS

var parsedTemplate = template.Must(
	template.ParseFS(templateFiles, "templates/email.html"),
)

type TemplateData struct {
	Body template.HTML
	Year int
	SupportEmail string
}

func Render(body string) (string, error) {
	if body == "" {
		return "", fmt.Errorf("email body is required")
	}

	var buf bytes.Buffer

	err := parsedTemplate.Execute(&buf, TemplateData{
		Body: template.HTML(body),
		Year: time.Now().Year(),
		SupportEmail: globals.SUPPORT_EMAIL,
	})
	if err != nil {
		return "", fmt.Errorf("render email: %w", err)
	}

	if buf.Len() == 0 {
		return "", fmt.Errorf("rendered email is empty")
	}

	return buf.String(), nil
}
