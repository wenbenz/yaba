.PHONY: clean build test docker ui dev dev-server dev-ui proto

build: deps graphql
	go build

deps:
	go get

graphql: graph/server/generated.go graph/model/models_gen.go

graph/server/generated.go: deps
	go run github.com/99designs/gqlgen generate

# Regenerate Go (gen/) and TS (ui/gen/, web/src/gen/) code from proto/. Requires the buf CLI.
proto:
	buf generate

clean:
	rm ./graph/client/generated.go ./graph/model/models_gen.go ./graph/server/generated.go \
		./yaba ./coverage.out

lint:
	golangci-lint run --fix

test:
	go test ./...
	cd ui && yarn test

cover:
	go test -v -race -covermode=atomic -coverprofile=coverage.out yaba/internal/... yaba/ui/...

docker:
	docker build --tag wenbenz/yaba:latest --tag wenbenz/yaba:$$(git rev-parse --short HEAD) .

# Build the UI into ui/dist. Run before `make build` to embed it in the binary.
ui:
	cd ui && yarn install --frozen-lockfile && yarn build

# Local dev: Postgres in Docker, Go server on :9222, Vite on :3000 with hot reload.
dev:
	docker compose up -d db
	$(MAKE) -j2 dev-server dev-ui

dev-server:
	env $$(grep -v '^#' docker.env | xargs) POSTGRES_URL=localhost:5432 \
		YABA_SERVER_PORT=9222 INSECURE_COOKIE=true go run .

dev-ui:
	cd ui && yarn install && yarn start
