package templating

import (
	"html/template"
	"io"
	"log/slog"
	"path/filepath"
)

type LiveLoader struct {
	Logger  *slog.Logger
	BaseDir string
}

func (l *LiveLoader) Render(w io.Writer, page string, data any) error {
	basePath := filepath.Join(l.BaseDir, "base.html")
	pagePath := filepath.Join(l.BaseDir, "pages", page)

	t, err := template.ParseFiles(basePath, pagePath)
	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, "main", data)
}
