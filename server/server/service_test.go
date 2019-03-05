package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const swaggerJSONPayload = `{
	"definitions": {
		"GiveNewSchemaNameHere": {
			"properties": {
				"amount": {
					"default": 1162,
					"type": "integer"
				},
				"balance": {
					"default": 718,
					"type": "integer"
				},
				"currency": {
					"default": "USD",
					"type": "string"
				},
				"id": {
					"default": "1234",
					"type": "string"
				},
				"ref": {
					"default": "INV-1234",
					"type": "string"
				}
			},
			"type": "object"
		}
	},
	"info": {
		"title": "invoiceservice",
		"version": "1.0.0",
		"x-lastModified": "Aug 08, 2018 13:35PM PST"
	},
	"paths": {
		"/api/invoices/{id}": {
			"get": {
				"operationId": "getApiInvoices_id",
				"parameters": [
					{
						"description": "",
						"format": "",
						"in": "path",
						"name": "id",
						"required": true,
						"type": "string"
					}
				],
				"produces": [
					"application/json"
				],
				"responses": {
					"200": {
						"description": "Success response",
						"examples": {
							"application/json": {
								"amount": 1162,
								"balance": 718,
								"currency": "USD",
								"id": "1234",
								"ref": "INV-1234"
							}
						},
						"schema": {
							"$ref": "#/definitions/GiveNewSchemaNameHere"
						}
					}
				}
			}
		}
	},
	"swagger": "2.0",
	"host": "localhost:8080",
	"schemes": [
		"http"
	],
	"basePath": "/"
}`

const kubeServicePayload = `{
    "metadata": {
        "name": "invoice-go-svc",
        "namespace": "default",
        "selfLink": "/api/v1/namespaces/default/services/invoice-go-svc",
        "uid": "a3f33d97-e0cb-11e8-8617-c85b76f2707f",
        "resourceVersion": "2982",
        "creationTimestamp": "2018-11-05T07:23:06Z",
        "labels": {
            "run": "invoice-go-svc"
        },
        "annotations": {
            "apiscout/index": "true",
            "apiscout/swaggerUrl": "/swaggerspec",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"v1\",\"kind\":\"Service\",\"metadata\":{\"annotations\":{\"apiscout/index\":\"true\",\"apiscout/swaggerUrl\":\"/swaggerspec\"},\"labels\":{\"run\":\"invoice-go-svc\"},\"name\":\"invoice-go-svc\",\"namespace\":\"default\"},\"spec\":{\"ports\":[{\"port\":80,\"protocol\":\"TCP\",\"targetPort\":8080}],\"selector\":{\"run\":\"invoice-go-svc\"},\"type\":\"LoadBalancer\"}}\n"
        }
    },
    "spec": {
        "ports": [
            {
                "protocol": "TCP",
                "port": 80,
                "targetPort": 8123,
                "nodePort": 8123
            }
        ],
        "selector": {
            "run": "invoice-go-svc"
        },
        "clusterIP": "10.99.164.156",
        "type": "LoadBalancer",
        "sessionAffinity": "None",
        "externalTrafficPolicy": "Cluster"
    },
    "status": {
        "loadBalancer": {}
    }
}`

func TestHandleService(t *testing.T) {
	server := &http.Server{Addr: ":8123"}
	http.HandleFunc("/swaggerspec", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, swaggerJSONPayload)
	})
	done := make(chan bool, 1)
	go func() {
		server.ListenAndServe()
		done <- true
	}()
	_, err := http.Get("http://localhost:8123/swaggerspec")
	for err != nil {
		_, err = http.Get("http://localhost:8123/swaggerspec")
	}
	defer func() {
		err := server.Shutdown(nil)
		if err != nil {
			t.Fatal(err)
		}
		<-done
	}()

	service := &v1.Service{}

	json.Unmarshal([]byte(kubeServicePayload), service)

	tempPath := "/tmp/apiscouttest1234"
	runMode := "LOCAL"
	externalIP := "localhost"

	os.MkdirAll(tempPath, 0777)

	srv, err := New(tempPath, tempPath, runMode, externalIP, tempPath, "", "")
	if err != nil {
		panic(err.Error())
	}

	srv.handleService(service, watch.Added, 0)
	if strings.Compare(srv.ServiceMap["invoice-go-svc"], "DONE") != 0 {
		t.Fatal("Service addition failed")
	}

	srv.handleService(service, watch.Deleted, 0)
	if len(srv.ServiceMap) != 0 {
		t.Fatal("Service removal failed")
	}

	os.RemoveAll(tempPath)

}
