FROM golang
WORKDIR /yaba
ENV GOPATH /yaba

# Download go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the app binary
COPY errors ./errors/
COPY config/*.go ./config/
COPY graph/ ./graph/
COPY internal ./internal/
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./yaba

# Unpack the UI
COPY dist.tar.gz ./
RUN tar -xzvf ./dist.tar.gz
ENV UI_ROOT_DIR /yaba/dist

# Open port
EXPOSE 80

# Start server
CMD ["./yaba"]
