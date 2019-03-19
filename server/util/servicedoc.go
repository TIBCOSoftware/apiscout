package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

const serviceMarkdown = `---
title: {{.title}}
weight: 5
---
`

// WriteServiceMarkdownFile writes service index file for given servicename
func WriteServiceMarkdownFile(path, serviceName string) error {

	fileName := filepath.Join(path, "_index.md")
	log.Printf("Preparing to write %s to disk", fileName)

	os.Remove(fileName)

	dataMap := make(map[string]interface{})
	dataMap["title"] = serviceName

	// Render the Markdown file based on the template
	t := template.Must(template.New("top").Parse(serviceMarkdown))
	buf := &bytes.Buffer{}

	if err := t.Execute(buf, dataMap); err != nil {
		log.Printf("error while rendering Markdown file: %s", err.Error())
		return fmt.Errorf("error while rendering Markdown file: %s", err.Error())
	}
	s := buf.String()

	s = s + "\n {{% children %}} "

	// creating file
	err := createFileWithContent(fileName, s)
	if err != nil {
		return err
	}

	return nil
}
