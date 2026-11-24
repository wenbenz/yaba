FROM node:lts AS ui-builder
WORKDIR /ui
COPY ui/package.json ui/yarn.lock ./
RUN yarn install --frozen-lockfile
COPY ui/ ./
RUN yarn build

FROM golang:1.26
WORKDIR /yaba
ENV GOPATH /yaba

# Download go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy migrations
COPY migrations ./migrations/

# Build the app binary
COPY errors ./errors/
COPY config/*.go ./config/
COPY graph/ ./graph/
COPY internal ./internal/
COPY ui/*.go ./ui/
COPY --from=ui-builder /ui/dist ./ui/dist/
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./yaba

# Open port
EXPOSE 80

# Start server
CMD ["./yaba"]
