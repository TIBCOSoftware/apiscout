// Package server implements the server of APIScout
package server

import (
	"log"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const maxRetryCount = 3

// retry takes a nap for 30 seconds in a separate go routine and retries the service. Usually when a service is created
// when the server component of the app isn't fully started (like on initial deployment), the server would respond with
// a dial timeout and it should be retried
func (srv *Server) retry(service *v1.Service, eventType watch.EventType, retryCount int) {
	if retryCount < maxRetryCount {
		go func() {
			log.Printf("Retrying %s in 30 seconds...", service.Name)
			time.Sleep(30000 * time.Millisecond)
			log.Printf("Retrying %s with current retryCount %d...", service.Name, retryCount)
			srv.handleService(service, eventType, retryCount+1)
		}()
	}
}
