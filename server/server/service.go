// Package server implements the server of APIScout
package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/TIBCOSoftware/apiscout/server/util"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	// The annotation for apiscout to index a service
	annotation = "apiscout/index"
	// The annotation for apiscout to get the OpenAPI doc from
	swaggerURL = "apiscout/swaggerUrl"

	publishToMasheryVal = "apiscout/publishToMashery"
	createPlanVal       = "masheryCreatePackagePlan"
	docTypeVal          = "masheryPublishDocType"
)

type Info struct {
	Info Title `json:"info"`
}

type Title struct {
	TitleVal string `json:"title"`
}

// handleService takes the Kubernetes service object and the EventType as input to determine what
// to do with the event
func (srv *Server) handleService(service *v1.Service, eventType watch.EventType, retryCount int) {
	log.Printf("Received %s for %s\n", eventType, service.Name)

	switch eventType {
	case watch.Added:
		if service.Annotations[annotation] == "true" {
			err := add(service, srv)
			if err != nil {
				if strings.Contains(err.Error(), "dial tcp") {
					srv.retry(service, eventType, retryCount+1)
				} else {
					log.Println(err.Error())
					return
				}
			}
		}
	case watch.Deleted:
		err := remove(service, srv)
		if err != nil {
			log.Println(err.Error())
			return
		}
	case watch.Modified:
		err := remove(service, srv)
		if err != nil {
			log.Println(err.Error())
		}
		if service.Annotations[annotation] == "true" {
			err := add(service, srv)
			if err != nil {
				if strings.Contains(err.Error(), "dial tcp") {
					srv.retry(service, eventType, retryCount+1)
				} else {
					log.Println(err.Error())
					return
				}
			}
		}
	case watch.Error:
		log.Println("Received watch.EventType Error, this is not recommended to be handled so API Scout will ignore")
		return
	default:
		log.Printf("Received unknown watch.EventType %s, so API Scout will ignore\n", eventType)
		return
	}

	// Generate the Hugo documentation
	err := util.GenerateDocs(srv.HugoDir)
	if err != nil {
		log.Printf("Error while attemtping to regenerate Hugo content: %s", err.Error())
	}
}

// add adds a service to the service map and generates the developer documentation if it doesn't exist in the service map yet
func add(service *v1.Service, srv *Server) error {
	if _, ok := srv.ServiceMap[service.Name]; !ok {
		log.Printf("%s should be indexed from %s\n", service.Name, service.Annotations[swaggerURL])

		var ip string
		var port int32

		if len(srv.ExternalIP) > 0 {
			ip = srv.ExternalIP
			port = service.Spec.Ports[0].NodePort
		} else {
			ip = service.Spec.ClusterIP
			port = service.Spec.Ports[0].Port
		}

		apidoc, err := util.GetAPIDoc(fmt.Sprintf("http://%s:%d%s", ip, port, service.Annotations[swaggerURL]))
		if err != nil {
			log.Printf("Error while retrieving API document from %s: %s", fmt.Sprintf("http://%s:%d%s", ip, port, service.Annotations[swaggerURL]), err.Error())
			return err
		}

		// Unmarshal the string into a proper document
		var swagger map[string]interface{}
		if err := json.Unmarshal([]byte(apidoc), &swagger); err != nil {
			log.Printf("error while unmarshaling JSON: %s", err.Error())
			return fmt.Errorf("error while unmarshaling JSON: %s", err.Error())
		}

		// Update the host information
		if _, ok := swagger["host"]; ok {
			swagger["host"] = fmt.Sprintf("%s:%d", ip, port)
		}

		var docTitle string
		if val, ok := swagger["info"].(map[string]interface{})["title"]; ok {
			docTitle = val.(string)
		}

		// Serialize the OpenAPI doc
		apibytes, err := json.Marshal(swagger)
		if err != nil {
			log.Printf("error while marshaling API: %s", err.Error())
			return fmt.Errorf("error while marshaling API: %s", err.Error())
		}

		util.WriteSwaggerToDisk(service.Name, apibytes, docTitle, srv.SwaggerStore, srv.HugoStore)

		srv.ServiceMap[service.Name] = "DONE"
		log.Printf("Service %s has been added to API Scout\n", service.Name)

		//////// PUBLISH TO MASHERY CODE ////////////
		if strings.Compare(strings.ToUpper(service.Annotations[publishToMasheryVal]), "TRUE") == 0 {

			docType := service.Annotations[docTypeVal]
			createPlan := false
			if strings.Compare(strings.ToUpper(service.Annotations[createPlanVal]), "TRUE") == 0 {
				createPlan = true
			}
			// setting default vaue if doc type not annotated in yml
			if len(docType) == 0 {
				docType = "IODOC"
			}

			//default api template
			apiTemplate := "/tmp/masheryTemplate.json"
			var apiTemplateJSON []byte
			_, err := os.Stat(apiTemplate)
			if err == nil {
				apiTemplateJSON, err = ioutil.ReadFile(apiTemplate)
				if err != nil {
					log.Fatal(err)
				}
			}

			user := APIUser{Username: srv.MasheryDetails.UserName, Password: srv.MasheryDetails.Password, APIKey: srv.MasheryDetails.APIKey, APISecretKey: srv.MasheryDetails.APISecret, UUID: srv.MasheryDetails.AreaID, Portal: srv.MasheryDetails.AreaDomain, Noop: false}

			err = PublishToMashery(&user, string(apibytes), docType, createPlan, apiTemplateJSON)
			if err != nil {
				log.Fatal(err)
			}

		}

	}

	return nil
}

// remove deletes the service from the service map and removes the JSON and Markdown files from disk
func remove(service *v1.Service, srv *Server) error {
	log.Printf("Attempting to delete %s\n", service.Name)

	// Remove JSON file
	filename := filepath.Join(srv.SwaggerStore, fmt.Sprintf("%s.json", strings.Replace(strings.ToLower(service.Name), " ", "-", -1)))

	apiDoc, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	f := &Info{}
	if err = json.Unmarshal(apiDoc, f); err != nil {
		panic(err)
	}

	err = os.Remove(filename)
	if err != nil {
		return err
	}

	// Remove Markdown file
	filename = filepath.Join(srv.HugoStore, fmt.Sprintf("%s.md", strings.Replace(strings.ToLower(service.Name), " ", "-", -1)))
	err = os.Remove(filename)
	if err != nil {
		return err
	}

	// Remove service from service map
	delete(srv.ServiceMap, service.Name)
	log.Printf("Service %s has been removed from API Scout\n", service.Name)

	user := APIUser{Username: srv.MasheryDetails.UserName, Password: srv.MasheryDetails.Password, APIKey: srv.MasheryDetails.APIKey, APISecretKey: srv.MasheryDetails.APISecret, UUID: srv.MasheryDetails.AreaID, Portal: srv.MasheryDetails.AreaDomain, Noop: false}

	RemoveFromMashery(&user, f.Info.TitleVal)

	return nil
}
