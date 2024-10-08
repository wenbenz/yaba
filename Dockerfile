FROM golang
WORKDIR /yaba
ENV GOPATH /yaba

# Download go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the app binary
COPY errors ./errors
COPY graph/ ./graph/
COPY internal ./internal
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./yaba

# Open port
EXPOSE 9222

# Start server
CMD ["./yaba"]
