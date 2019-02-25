REGISTRY := sevein
IMAGE := $(REGISTRY)/k8s-sysdig-adapter
VERSION := $(shell git describe --tags --always --dirty)

test: install
	go test ./...

check: test
	@echo Checking code is gofmted
	@bash -c 'if [ -n "$(gofmt -s -l .)" ]; then echo "Go code is not formatted:"; gofmt -s -d -e .; exit 1;fi'

install:
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go install -v -ldflags="-w -s" -v github.com/draios/kubernetes-sysdig-metrics-apiserver/cmd/adapter 

build-image:
	docker build -t $(IMAGE):$(VERSION) .
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

push-image: build-image
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest
