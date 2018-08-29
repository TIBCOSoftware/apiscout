.PHONY: clean-all clean-docker clean-kube build-site build-server build-docker build-all run-server run-docker run-hugo run-kube

#--- Variables ---
CURRDIR = `pwd`
EXTIP = `minikube ip`
USER = `whoami`
DOCKERREPO = retgits

#--- Help ---
help:
	@echo Makefile for apiscout
	@echo  
	@echo usage: make [target]
	@echo
	@echo Cleaning targets:
	@echo - clean-all : Removes the dist directory
	@echo - clean-docker : Stops and removes all containers and images for apiscout
	@echo - clean-kube : Removes the apiscout service and deployment from Kubernetes
	@echo
	@echo Build targets:
	@echo - build-site : Builds the Hugo distribution in dist
	@echo - build-server : Builds the Flogo server in dist
	@echo - build-docker : Builds a docker image from the dist directory
	@echo - build-all : Performs clean-all and executes all build targets
	@echo
	@echo Run targets
	@echo - run-server : Builds the  in the server directory and runs it with default settings
	@echo - run-docker : Runs a docker container with default settings
	@echo - run-hugo : Runs the embedded Hugo server on port 1313
	@echo - run-kube : Deploys apiscout to Kubernetes
	@echo
	@echo Stop targets
	@echo - stop-docker : Stop and remove the running apiscout container

#--- Clean up targets ---
clean-all:
	rm -rf ./dist
	
clean-docker:
	docker stop apiscout
	docker rm apiscout
	docker rmi retgits/apiscout

clean-kube:
	kubectl delete svc apiscout-svc
	kubectl delete deployment apiscout

#--- Build targets ---
build-site:
	mkdir -p dist
	cp -r webapp/* ./dist

build-server:
	mkdir -p dist
	cd server && go generate && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../dist/server *.go

build-docker:
	cp Dockerfile ./dist/Dockerfile
	cp -R ./nginx/ ./dist/nginx
	cd dist && docker build . -t $(DOCKERREPO)/apiscout:latest

build-all: clean-all build-site build-server build-docker

#--- Run targets ---
run-server:
	cd server && go generate && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../dist/server *.go
	MODE=LOCAL HUGODIR=$(CURRDIR)/webapp HUGOSTORE=$(CURRDIR)/webapp/content/apis SWAGGERSTORE=$(CURRDIR)/webapp/static/swaggerdocs EXTERNALIP=$(EXTIP) ./dist/server

run-docker:
	docker run -it --rm -p 80:80 -v $(HOME)/.kube:/root/.kube -v $(HOME)/.minikube:/home/$(USER)/.minikube -e MODE=LOCAL -e HUGODIR="/tmp" -e EXTERNALIP=$(EXTIP) -e HUGOCMD="sh -c \"cd /tmp && hugo\"" --name=apiscout $(DOCKERREPO)/apiscout:latest

run-hugo:
	cd webapp && hugo server -D --disableFastRender

run-kube:
	kubectl apply -f ./kubernetes/apiscout.yml

#--- Stop targets ---
stop-docker:
	docker stop apiscout && docker rm apiscout

#--- Minikube targets ---
minikube-install:
	curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube
minikube-start:
	export MINIKUBE_WANTUPDATENOTIFICATION=false
	export MINIKUBE_WANTREPORTERRORPROMPT=false
	export MINIKUBE_HOME=$(HOME)
	export CHANGE_MINIKUBE_NONE_USER=true
	export KUBECONFIG=$(HOME)/.kube/config
	sudo -E minikube start --vm-driver=none
minikube-stop:
	minikube stop
minikube-delete: minikube-stop
	minikube delete
minikube-show:
	open `minikube service apiscout-svc --url`