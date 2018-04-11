REGISTRY := sevein
IMAGE := $(REGISTRY)/k8s-sysdig-adapter
VERSION := $(shell git describe --tags --always --dirty)

test: install
	go test ./...

check: test
	@echo Checking code is gofmted
	@bash -c 'if [ -n "$(gofmt -s -l .)" ]; then echo "Go code is not formatted:"; gofmt -s -d -e .; exit 1;fi'

install:
	go install -v ./...

build-image:
	docker build -t $(IMAGE):$(VERSION) .

push-image: build-image
	docker push $(IMAGE):$(VERSION)
