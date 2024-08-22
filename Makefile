.PHONY: clean deps docker graphql
yaba: deps graphql build

deps:
	go get

graphql: internal/handlers/generated.go graph/model/models_gen.go graph/client/generated.go

graph/client/generated.go:
	go run github.com/Khan/genqlient genqlient.yaml

graph/model/models_gen.go: internal/handlers/generated.go

internal/handlers/generated.go:
	go run github.com/99designs/gqlgen generate

build:
	go build

clean:
	rm ./graph/client/generated.go ./graph/model/models_gen.go ./internal/handlers/generated.go \
		./yaba ./coverage.out

docker:
	docker build --tag wenbenz/yaba:latest .
