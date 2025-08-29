package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"golang.org/x/net/html"
)

func ExtractContent(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	switch ext {
	case ".txt":
		return extractTextFile(file)
	case ".pdf":
		return extractPDFContent(file)
	case ".html", ".htm":
		return extractHTMLContent(file)
	case ".json":
		return extractJSONContent(file)
	default:
		return extractTextFile(file)
	}
}

func extractPDFContent(file multipart.File) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(data)
	f, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", err
	}

	var content strings.Builder
	maxPages := 10

	for i := 1; i <= min(f.NumPage(), maxPages); i++ {
		page := f.Page(i)
		text, err := page.GetPlainText(make(map[string]*pdf.Font))
		if err != nil {
			continue
		}
		content.WriteString(text)
		content.WriteString("\n")
	}

	return content.String(), nil
}

func extractHTMLContent(file multipart.File) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	var content strings.Builder
	extractTextFromHTML(doc, &content)

	return content.String(), nil
}

func extractTextFromHTML(node *html.Node, content *strings.Builder) {
	if node.Type == html.TextNode {
		text := strings.TrimSpace(node.Data)
		if text != "" {
			content.WriteString(text)
			content.WriteString(" ")
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		extractTextFromHTML(child, content)
	}
}

func extractJSONContent(file multipart.File) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return string(data), nil
	}

	formatted, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return string(data), nil
	}

	return string(formatted), nil
}

func extractTextFile(file multipart.File) (string, error) {
	const maxChar = 50000
	buf := make([]byte, maxChar)

	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(buf[:n]), nil
}
