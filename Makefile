.PHONY: clean deps docker graphql

build: deps graphql
	go build

deps:
	go get

graphql: internal/handlers/generated.go graph/model/models_gen.go graph/client/generated.go

graph/model/models_gen.go: internal/handlers/generated.go

graph/client/generated.go: deps
	go run github.com/Khan/genqlient genqlient.yaml

internal/handlers/generated.go: deps
	go run github.com/99designs/gqlgen generate

clean:
	rm ./graph/client/generated.go ./graph/model/models_gen.go ./internal/handlers/generated.go \
		./yaba ./coverage.out

docker:
	docker build --tag wenbenz/yaba:latest .

cover:
	go test -v -race -covermode=atomic -coverprofile=coverage.out yaba/internal/...