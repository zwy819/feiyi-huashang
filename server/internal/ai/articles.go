package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Article struct {
	Slug      string
	Title     string
	Publisher string
	SourceURL string
	License   string
	Accessed  string
	Tags      string
	Content   string
}

// LoadArticles reads the small, human-curated knowledge base shipped with the
// server. Front matter is intentionally parsed as simple key: value pairs so
// the corpus has no YAML dependency and remains easy to review in Git.
func LoadArticles(dir string) ([]Article, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	articles := make([]Article, 0, len(paths))
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		article, err := ParseArticle(string(raw))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		article.Slug = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if article.Title == "" {
			return nil, fmt.Errorf("%s: missing title", filepath.Base(path))
		}
		if article.SourceURL == "" {
			return nil, fmt.Errorf("%s: missing source_url", filepath.Base(path))
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func ParseArticle(raw string) (Article, error) {
	raw = strings.TrimPrefix(raw, "\ufeff")
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	article := Article{}
	if !strings.HasPrefix(raw, "---\n") {
		return article, fmt.Errorf("missing front matter")
	}
	rest := strings.TrimPrefix(raw, "---\n")
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return article, fmt.Errorf("unterminated front matter")
	}
	meta := rest[:end]
	article.Content = strings.TrimSpace(rest[end+5:])
	if article.Content == "" {
		return article, fmt.Errorf("empty content")
	}

	for _, line := range strings.Split(meta, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "title":
			article.Title = value
		case "publisher":
			article.Publisher = value
		case "source_url", "url":
			article.SourceURL = value
		case "license":
			article.License = value
		case "accessed":
			article.Accessed = value
		case "tags":
			article.Tags = value
		}
	}
	return article, nil
}
