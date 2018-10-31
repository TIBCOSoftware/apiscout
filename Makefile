.PHONY: clean-all clean-docker clean-kube build-site build-server build-docker build-all run-server run-docker run-hugo run-kube

#--- Variables ---
CURRDIR = `pwd`
EXTIP = `minikube ip`
USER = `whoami`
DOCKERREPO = `whoami`
KUBEFILES = .

#--- Help ---
help:
	@echo 
	@echo Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo

#--- Clean up targets ---
clean-all: ## Removes the dist directory
	rm -rf ./dist
	
clean-docker: ## Stops and removes all containers and images for apiscout
	docker stop apiscout
	docker rm apiscout
	docker rmi retgits/apiscout

clean-kube: ## Removes the apiscout service and deployment from Kubernetes
	kubectl delete svc apiscout-svc
	kubectl delete deployment apiscout

#--- Get dependencies ---
deps: ## Get dependencies to build the server
	go get -u github.com/TIBCOSoftware/apiscout/server

#--- Build targets ---
build-site: ## Builds the Hugo distribution in dist
	mkdir -p dist
	cp -r webapp/* ./dist

build-server: ## Builds the server app in dist
	mkdir -p dist
	cd server && go generate && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../dist/server *.go

build-docker: ## Builds a docker image from the dist directory
	cp Dockerfile ./dist/Dockerfile
	cp -R ./nginx/ ./dist/nginx
	cd dist && docker build . -t $(DOCKERREPO)/apiscout:latest

build-all: clean-all build-site build-server build-docker ## Performs clean-all and executes all build targets

#--- Run targets ---
run-server: ## Builds the  in the server directory and runs it with default settings
	cd server && go generate && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../dist/server *.go
	MODE=LOCAL HUGODIR=$(CURRDIR)/webapp HUGOSTORE=$(CURRDIR)/webapp/content/apis SWAGGERSTORE=$(CURRDIR)/webapp/static/swaggerdocs EXTERNALIP=$(EXTIP) ./dist/server

run-docker: ## Runs a docker container with default settings
	docker run -it --rm -p 80:80 -v $(HOME)/.kube:/root/.kube -v $(HOME)/.minikube:/home/$(USER)/.minikube -e MODE=LOCAL -e HUGODIR="/tmp" -e EXTERNALIP=$(EXTIP) -e HUGOCMD="sh -c \"cd /tmp && hugo\"" --name=apiscout $(DOCKERREPO)/apiscout:latest

run-hugo: ## Runs the embedded Hugo server on port 1313
	cd webapp && hugo server -D --disableFastRender

run-docs: ## Runs the embedded Hugo server on port 1313 for the documentation
	cd docs && hugo server -D --disableFastRender --themesDir ../webapp/themes

run-kube: ## Deploys apiscout to Kubernetes
	kubectl apply -f ${KUBEFILES}/apiscout.yml

#--- Stop targets ---
stop-docker: ## Stop and remove the running apiscout container
	docker stop apiscout && docker rm apiscout

#--- Minikube targets ---
minikube-install: ## Install Minikube on this machine
	curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube
minikube-start: ## Start Minikube with default configuration
	export MINIKUBE_WANTUPDATENOTIFICATION=false
	export MINIKUBE_WANTREPORTERRORPROMPT=false
	export MINIKUBE_HOME=$(HOME)
	export CHANGE_MINIKUBE_NONE_USER=true
	export KUBECONFIG=$(HOME)/.kube/config
	sudo -E minikube start --vm-driver=none
minikube-stop: ## Stop Minikube
	minikube stop
minikube-delete: minikube-stop ## Delete the Minikube installation
	minikube delete
minikube-show: ## Show the API Scout UI that is deployed to Minikube
	open `minikube service apiscout-svc --url`
