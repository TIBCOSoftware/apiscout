.PHONY: deps

#--- Variables ---
DOCKERHUBUSER=naresh
DOCKERTAG=latest
IPADDR=xxx.xxx.xxx.xxx

#--- Get the dependencies ---
deps: 
	go get -u ./...

#--- Clean up the dist folder ---
clean:
	rm -rf dist

#--- Clean up the Kubernetes deployment ---
clean-kube:
	kubectl delete deployment streetlights-go-svc
	kubectl delete svc streetlights-go-svc

#--- Build the Flogo app ---
build-app:
	go generate
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/streetlightsapp-go

#--- Build the docker container ---
build-docker:
	cp Dockerfile dist/Dockerfile
	cp streetlightsapi.yaml dist/streetlightsapi.yaml
	eval $$(minikube docker-env) ;\
	cd dist && docker build -t $(DOCKERHUBUSER)/streetlightsapp-go:$(DOCKERTAG) .

#--- Run the container ---
run-docker:
	docker run --rm -it -p 9999:8090 $(DOCKERHUBUSER)/streetlightsapp-go:$(DOCKERTAG)

#--- Run the app on Kubernetes ---
run-kube:
	kubectl apply -f streetlights-go-svc.yml