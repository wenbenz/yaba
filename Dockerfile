FROM golang
WORKDIR /yaba
ENV GOPATH /yaba

# Download go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the app binary
COPY internal ./internal
COPY server ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./yaba

# Open port
EXPOSE 8080

# Start server
CMD ["./yaba"]
