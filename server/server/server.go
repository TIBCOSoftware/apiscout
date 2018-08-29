// Package server implements the server of APIScout
package server

import (
	"path/filepath"
	"strings"

	"github.com/TIBCOSoftware/apiscout/server/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Server represents the APIScout server and implements methods.
type Server struct {
	// A map[string]string of all services that have been indexed by apiscout
	ServiceMap map[string]string
	// The location where to store the swaggerdocs
	SwaggerStore string
	// The location where to store content for Hugo
	HugoStore string
	// The mode in which apiscout is running (can be either KUBE or LOCAL)
	RunMode string
	// The external IP address of the Kubernetes cluster in case of LOCAL mode
	ExternalIP string
	// The base directory for Hugo
	HugoDir string
}

// New creates a new instance of the Server
func New(swaggerStore string, hugoStore string, runMode string, externalIP string, hugoDir string) (*Server, error) {
	// Return a new struct
	return &Server{
		ServiceMap:   make(map[string]string),
		SwaggerStore: swaggerStore,
		HugoStore:    hugoStore,
		RunMode:      runMode,
		ExternalIP:   externalIP,
		HugoDir:      hugoDir,
	}, nil
}

// Start is the main engine to start the APIScout server
func (srv *Server) Start() {
	var config *rest.Config
	var err error

	if strings.ToUpper(srv.RunMode) == "KUBE" {
		// Create the Kubernetes in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(util.HomeDir(), ".kube", "config"))
		if err != nil {
			panic(err.Error())
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create a watcher
	watcher, err := clientset.CoreV1().Services("").Watch(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Create a channel for the events to come in from the watcher
	eventChannel := watcher.ResultChan()

	// Start an indefinite loop
	for {
		evt := <-eventChannel
		srv.handleEvent(evt)
	}
}
