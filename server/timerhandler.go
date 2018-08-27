//go:generate go run $GOPATH/src/github.com/TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH

// Package main
package main

// The imports
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// timerHandler handles the recurring timer events that are triggered by the Flogo engine
func timerHandler(ctx context.Context, inputs map[string]*data.Attribute) (map[string]*data.Attribute, error) {
	var config *rest.Config
	var err error

	if strings.ToUpper(runMode) == "KUBE" {
		// Create the Kubernetes in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homeDir(), ".kube", "config"))
		if err != nil {
			panic(err.Error())
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// get all the services
	services, err := clientset.CoreV1().Services("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	log.Printf("A total of %d services have been found to index\n", len(services.Items))

	regenerateHugo := false

	// find the services that need to be indexed
	for _, service := range services.Items {
		if service.Annotations[apiscoutAnnotation] == "true" {
			// Only index if it hasn't been indexed before
			if _, ok := serviceMap[service.Name]; !ok {
				regenerateHugo = true
				fmt.Printf("%s should be indexed from %s\n", service.Name, service.Annotations[apiscoutSwaggerURL])
				var ip string
				var port int32
				if strings.ToUpper(runMode) == "KUBE" {
					ip = service.Spec.ClusterIP
					port = service.Spec.Ports[0].Port
				} else {
					ip = externalIP
					port = service.Spec.Ports[0].NodePort
				}

				apidoc, err := getAPIDoc(fmt.Sprintf("http://%s:%d%s", ip, port, service.Annotations[apiscoutSwaggerURL]))
				if err != nil {
					return nil, err
				}
				writeSwaggerToDisk(service.Name, apidoc, fmt.Sprintf("%s:%d", ip, port))
				serviceMap[service.Name] = "DONE"
			}
		}
	}

	if regenerateHugo && len(hugoDir) > 0 {
		log.Print("Regeneratig Hugo content")
		cmd := exec.Command("sh", "-c", "hugo")
		cmd.Dir = hugoDir
		log.Printf("%s", cmd.Args)
		output, err := cmd.CombinedOutput()
		log.Print(string(output))
		if err != nil {
			log.Printf("error while regeneratig Hugo content: %s", err.Error())
			return nil, fmt.Errorf("error while regeneratig Hugo content: %s", err.Error())
		}
	} else {
		log.Print("No new services were added to apiscout")
	}

	// Nothing to return
	return nil, nil
}

// getAPIDoc performs an HTTP request to a specified URL to retrieve the OpenAPI document
func getAPIDoc(url string) (string, error) {
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

func writeSwaggerToDisk(name string, apidoc string, svchost string) error {
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
	storageLocation := make([]string, 0)
	storageLocation = append(storageLocation, "/")
	storageLocation = append(storageLocation, strings.Split(swaggerStore, "/")...)
	storageLocation = append(storageLocation, fmt.Sprintf("%s.json", strings.Replace(strings.ToLower(name), " ", "-", -1)))
	filename := filepath.Join(storageLocation...)
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
	var description, title string
	if val, ok := swagger["info"].(map[string]interface{})["description"]; ok {
		description = val.(string)
	}
	if val, ok := swagger["info"].(map[string]interface{})["title"]; ok {
		title = val.(string)
	}

	dataMap := make(map[string]interface{})
	dataMap["description"] = description
	dataMap["title"] = title
	dataMap["json"] = fmt.Sprintf("{{< oai-spec url=\"../../../swaggerdocs/%s.json\" >}}", strings.Replace(strings.ToLower(name), " ", "-", -1))

	// Render the Markdown file based on the template
	t := template.Must(template.New("top").Parse(markdown))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, dataMap); err != nil {

	}
	s := buf.String()

	// Determine where to save the file
	storageLocation = make([]string, 0)
	storageLocation = append(storageLocation, "/")
	storageLocation = append(storageLocation, strings.Split(hugoStore, "/")...)
	storageLocation = append(storageLocation, fmt.Sprintf("%s.md", strings.Replace(strings.ToLower(name), " ", "-", -1)))
	filename = filepath.Join(storageLocation...)
	log.Printf("Preparing to write %s to disk", filename)
	os.Remove(filename)

	// Create a file on disk
	file, err = os.Create(filename)
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

	// Write the Markdown doc to disk
	_, err = file.Write([]byte(s))
	if err != nil {
		log.Printf("error while writing Markdown to disk: %s", err.Error())
		return fmt.Errorf("error while writing Markdown to disk: %s", err.Error())
	}

	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
