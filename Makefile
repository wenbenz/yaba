.PHONY: clean build test docker

build: deps graphql
	go build

deps:
	go get

graphql: graph/server/generated.go graph/model/models_gen.go

graph/server/generated.go: deps
	go run github.com/99designs/gqlgen generate

clean:
	rm ./graph/client/generated.go ./graph/model/models_gen.go ./graph/server/generated.go \
		./yaba ./coverage.out

lint:
	golangci-lint run --fix 

test:
	go test ./...

cover:
	go test -v -race -covermode=atomic -coverprofile=coverage.out yaba/internal/...

docker:
	docker build --tag wenbenz/yaba:latest .
