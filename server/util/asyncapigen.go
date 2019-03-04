package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// A template for the Markdown file for Hugo
const asyncMarkdown = `---
title: {{.title}}
weight: 1000
---

{{.content}}`

// GenerateMarkdownFile generates markdown file for asyncapi docs
func GenerateMarkdownFile(srcFile, destFile, apiDoc, serviceName string) error {

	log.Print("Writing json file from received data")

	asyncFile := filepath.Join(srcFile, fmt.Sprintf("%s.json", strings.Replace(strings.ToLower(serviceName), " ", "-", -1)))
	log.Printf("Preparing to write %s to disk", asyncFile)
	os.Remove(asyncFile)

	err := createFileWithContent(asyncFile, apiDoc)
	if err != nil {
		return err
	}

	log.Print("Generating AsyncApi markdown file")

	cmd := exec.Command("ag", "-o", destFile, asyncFile, "markdown")
	output, err := cmd.CombinedOutput()
	log.Print(string(output))

	if err != nil {
		return err
	}

	// read markdown file
	data, err := ioutil.ReadFile(filepath.Join(destFile, "asyncapi.md"))
	if err != nil {
		return nil
	}
	os.Remove(filepath.Join(destFile, "asyncapi.md"))

	// Render the Markdown file based on the template
	t := template.Must(template.New("top").Parse(asyncMarkdown))
	buf := &bytes.Buffer{}

	dataMap := make(map[string]interface{})
	dataMap["title"] = "testingapidoc"
	dataMap["content"] = string(data)

	if err := t.Execute(buf, dataMap); err != nil {
		log.Printf("error while rendering Markdown file: %s", err.Error())
		return fmt.Errorf("error while rendering Markdown file: %s", err.Error())
	}
	s := buf.String()

	// Determine where to save the file
	mdFilename := filepath.Join(destFile, fmt.Sprintf("%s.md", strings.Replace(strings.ToLower(serviceName), " ", "-", -1)))
	log.Printf("Preparing to write %s to disk", mdFilename)
	os.Remove(mdFilename)

	err = createFileWithContent(mdFilename, s)
	if err != nil {
		return err
	}

	return nil
}
