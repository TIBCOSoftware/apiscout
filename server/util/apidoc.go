// Package util implements utility methods
package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// A template for the Markdown file for Hugo
const markdown = `---
title: {{.title}}
weight: 1000
post: "<sup><i>openapi</i></sup>"
---

{{.json}}`

// GetAPIDoc performs an HTTP request to a specified URL to retrieve the OpenAPI document
func GetAPIDoc(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// WriteSwaggerToDisk takes a swagger document and writes both its content as well as a hugo template to disk
// to enable the static site to be updated with the new API
func WriteSwaggerToDisk(name string, apidoc string, svchost string, swaggerStore string, hugoStore string) error {
	// Unmarshal the string into a proper document
	var swagger map[string]interface{}
	if err := json.Unmarshal([]byte(apidoc), &swagger); err != nil {
		log.Printf("error while unmarshaling JSON: %s", err.Error())
		return fmt.Errorf("error while unmarshaling JSON: %s", err.Error())
	}

	// Update the host information
	if _, ok := swagger["host"]; ok {
		swagger["host"] = svchost
	}

	// Determine where to save the file
	filename := filepath.Join(swaggerStore, fmt.Sprintf("%s.json", strings.Replace(strings.ToLower(name), " ", "-", -1)))
	log.Printf("Preparing to write %s to disk", filename)
	os.Remove(filename)

	// Create a file on disk
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("error while creating file: %s", err.Error())
		return fmt.Errorf("error while creating file: %s", err.Error())
	}
	defer file.Close()

	// Open the file to write
	file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("error while opening file: %s", err.Error())
		return fmt.Errorf("error while opening file: %s", err.Error())
	}

	// Serialize the OpenAPI doc
	apibytes, err := json.Marshal(swagger)
	if err != nil {
		log.Printf("error while marshaling API: %s", err.Error())
		return fmt.Errorf("error while marshaling API: %s", err.Error())
	}

	// Write the OpenAPI doc to disk
	_, err = file.Write(apibytes)
	if err != nil {
		log.Printf("error while writing OpenAPI to disk: %s", err.Error())
		return fmt.Errorf("error while writing OpenAPI to disk: %s", err.Error())
	}

	// Prepare the Markdown file for Hugo
	var title string
	if val, ok := swagger["info"].(map[string]interface{})["title"]; ok {
		title = val.(string)
	}

	dataMap := make(map[string]interface{})
	dataMap["title"] = title
	dataMap["json"] = fmt.Sprintf("{{< oas3 url=\"../../../swaggerdocs/%s.json\" >}}", strings.Replace(strings.ToLower(name), " ", "-", -1))

	// Render the Markdown file based on the template
	t := template.Must(template.New("top").Parse(markdown))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, dataMap); err != nil {
		log.Printf("error while rendering Markdown file: %s", err.Error())
		return fmt.Errorf("error while rendering Markdown file: %s", err.Error())
	}
	s := buf.String()

	// Determine where to save the file
	filename = filepath.Join(hugoStore, fmt.Sprintf("%s-openapi.md", strings.Replace(strings.ToLower(name), " ", "-", -1)))
	log.Printf("Preparing to write %s to disk", filename)
	os.Remove(filename)

	err = createFileWithContent(filename, s)
	if err != nil {
		return err
	}

	return nil
}
