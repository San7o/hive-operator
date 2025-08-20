
REGISTRY_NAME=giovann103

.PHONY: build
build:
	docker build -t ${REGISTRY_NAME}/ebpfdev:latest . -f Dockerfile.ebpfdev

.PHONY: push
push:
	docker push ${REGISTRY_NAME}/ebpfdev:latest
