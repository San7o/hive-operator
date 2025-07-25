IMG?=localhost:5001/callback:latest

.PHONY: build run test
all: build

build:
	go build -o callback cmd/main.go

run:
	./callback

test:
	curl -X POST -d "test" http://localhost:8090/ingest

.PHONY: deploy-local kill-pod-local
deploy:	
	kubectl apply -f config/

kill-pod:
	sudo kubectl delete pod callback-pod

.PHONY: docker-build docker-push docker
docker-build:
	docker rmi callback ${IMG} &2>/dev/null || :
	docker build -t callback .
	docker tag callback ${IMG}

docker-push:
	docker push ${IMG}

docker: docker-build docker-push
