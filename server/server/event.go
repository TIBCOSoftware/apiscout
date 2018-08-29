// Package server implements the server of APIScout
package server

import (
	"reflect"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// handleEvent takes care of the event itself and forwards the event in case it is a service
func (srv *Server) handleEvent(event watch.Event) {
	v := &v1.Service{}
	if reflect.TypeOf(event.Object) == reflect.TypeOf(v) {
		i := reflect.ValueOf(event.Object).Interface()
		srv.handleService(i.(*v1.Service), event.Type)
	}
}
