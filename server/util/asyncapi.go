package util

import (
	"bytes"
	"encoding/json"
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
title: {{.title}}-asyncapi
weight: 1000
#post: "<sup><i>asyncapi</i></sup>"
---

{{.content}}`

// GenerateMarkdownFile generates markdown file for asyncapi docs
func GenerateMarkdownFile(srcFile, destFile, apiDoc, serviceName string) error {

	log.Print("Writing file from received data")

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

	// Unmarshal the string into a proper document
	var docContent map[string]interface{}
	if err := json.Unmarshal([]byte(apiDoc), &docContent); err != nil {
		log.Printf("error while unmarshaling apidoc: %s", err.Error())
		return fmt.Errorf("error while unmarshaling apidoc: %s", err.Error())
	}

	var title string
	if val, ok := docContent["info"].(map[string]interface{})["title"]; ok {
		title = val.(string)
	}

	dataMap := make(map[string]interface{})
	dataMap["title"] = title
	dataMap["content"] = string(data)

	if err := t.Execute(buf, dataMap); err != nil {
		log.Printf("error while rendering Markdown file: %s", err.Error())
		return fmt.Errorf("error while rendering Markdown file: %s", err.Error())
	}
	s := buf.String()

	// Determine where to save the file
	mdFilename := filepath.Join(destFile, strings.Replace(strings.ToLower(serviceName), " ", "-", -1), fmt.Sprintf("%s-asyncapi.md", strings.Replace(strings.ToLower(serviceName), " ", "-", -1)))
	log.Printf("Preparing to write %s to disk", mdFilename)
	os.Remove(mdFilename)

	err = createFileWithContent(mdFilename, s)
	if err != nil {
		return err
	}

	return nil
}
