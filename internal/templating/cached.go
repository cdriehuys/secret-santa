package templating

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
)

type TemplateCache struct {
	logger *slog.Logger

	cache map[string]*template.Template
}

func NewTemplateCache(logger *slog.Logger, files fs.FS) (*TemplateCache, error) {
	basePath := "base.html"
	pagesPath := "pages"

	var pages []string
	visit := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".html" {
			pages = append(pages, path)
		}

		return nil
	}

	if err := fs.WalkDir(files, pagesPath, visit); err != nil {
		return nil, fmt.Errorf("failed to collect pages: %v", err)
	}

	cache := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		name, err := filepath.Rel(pagesPath, page)
		if err != nil {
			return nil, fmt.Errorf("failed to determine relative path for page %q: %v", page, err)
		}

		t, err := template.ParseFS(files, basePath, page)
		if err != nil {
			return nil, fmt.Errorf("failed to construct template for page %q: %v", page, err)
		}

		cache[name] = t
	}

	return &TemplateCache{logger, cache}, nil
}

func (c *TemplateCache) Render(w io.Writer, page string, data any) error {
	t, exists := c.cache[page]
	if !exists {
		return fmt.Errorf("template not found: %v", page)
	}

	return t.ExecuteTemplate(w, "main", data)
}
