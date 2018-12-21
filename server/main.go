// Package main implements the main start of API Scout
//
// API Scout helps you catalog and document your Kubernetes microservices so you know
// what you've deployed last summer!
//
// API Scout automatically discover microservices by using annotations
// * apiscout/index: This annotation ensures that apiscout indexes the service
// * apiscout/swaggerUrl: This is the URL from where apiscout will read the OpenAPI document
//
// After discovery, API Scout generates pixel-perfect OAS/Swagger-based API Docs and
// displays it using a staticly generated site (powered by [Hugo](https://gohugo.io))
package main

// The imports
import (
	"log"

	"github.com/TIBCOSoftware/apiscout/server/server"
	"github.com/TIBCOSoftware/apiscout/server/util"
)

const (
	// Version
	version = "0.2.0"
)

var (
	// The location where to store the swaggerdocs
	swaggerStore = util.GetEnvKey("SWAGGERSTORE", "/tmp/static/swaggerdocs")
	// The location where to store content for Hugo
	hugoStore = util.GetEnvKey("HUGOSTORE", "/tmp/content/apis")
	// The mode in which apiscout is running (can be either KUBE or LOCAL)
	runMode = util.GetEnvKey("MODE", "LOCAL")
	// The external IP address of the Kubernetes cluster in case of LOCAL mode
	externalIP = util.GetEnvKey("EXTERNALIP", "")
	// The base directory for Hugo
	hugoDir = util.GetEnvKey("HUGODIR", "")
)

// main is the main entrypoint to start APIScout
func main() {
	// Print config
	log.Printf("------------------------------------------------------------\n")
	log.Printf("CONFIG\n")
	log.Printf("APIScout version : %s\n", version)
	log.Printf("Run mode         : %s\n", runMode)
	log.Printf("Swagger store    : %s\n", swaggerStore)
	log.Printf("Hugo store       : %s\n", hugoStore)
	if len(externalIP) > 0 {
		log.Printf("External IP      : %s\n", externalIP)
	}
	if len(hugoDir) > 0 {
		log.Printf("Hugo dir         : %s\n", hugoDir)
	}
	log.Printf("------------------------------------------------------------\n")

	mshryInfo := server.MasheryInfo{
		UserName:   util.GetEnvKey("USERNAME", ""),
		Password:   util.GetEnvKey("PASSWORD", ""),
		APIKey:     util.GetEnvKey("APIKEY", ""),
		APISecret:  util.GetEnvKey("APISECRETE", ""),
		AreaID:     util.GetEnvKey("AREAID", ""),
		AreaDomain: util.GetEnvKey("AREADOMAIN", ""),
	}

	// Create a new APIScout server instance
	srv, err := server.New(swaggerStore, hugoStore, runMode, externalIP, hugoDir, mshryInfo)
	if err != nil {
		panic(err.Error())
	}

	// Start APIScout server
	srv.Start()
}
