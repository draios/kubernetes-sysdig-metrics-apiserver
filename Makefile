REGISTRY := registry.ng.bluemix.net/berg
IMAGE := $(REGISTRY)/k8s-sysdig-adapter
VERSION := $(shell git describe --tags --always --dirty)

test: install
	go test ./...

check: test
	@echo Checking code is gofmted
	@bash -c 'if [ -n "$(gofmt -s -l .)" ]; then echo "Go code is not formatted:"; gofmt -s -d -e .; exit 1;fi'

install:
	CGO_ENABLED=0 GOOS=linux go install -v ./...

build-image:
	docker build --no-cache -t $(IMAGE):$(VERSION) .
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

push-image: build-image
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest
