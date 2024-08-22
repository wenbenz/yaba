.PHONY: clean deps gqlgen gqlient docker
yaba: deps graphql build

deps:
	go get

graphql: gqlgen gqlient

gqlient:
	go run github.com/Khan/genqlient genqlient.yaml

gqlgen:
	go run github.com/99designs/gqlgen generate

build:
	go build

clean:
	rm ./internal/graph/client/generated.go ./internal/graph/model/models_gen.go ./internal/graph/server/generated.go \
		./yaba ./coverage.out

docker:
	docker build --tag wenbenz/yaba:latest .
