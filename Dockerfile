FROM node:lts AS ui-builder
WORKDIR /ui
COPY ui/package.json ui/yarn.lock ./
RUN yarn install --frozen-lockfile
COPY ui/ ./
RUN yarn build

FROM golang
WORKDIR /yaba
ENV GOPATH /yaba

# Download go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy migrations
COPY migrations ./migrations/

# Copy UI build output
COPY --from=ui-builder /ui/dist ./dist
ENV UI_ROOT_DIR /yaba/dist

# Build the app binary
COPY errors ./errors/
COPY config/*.go ./config/
COPY graph/ ./graph/
COPY internal ./internal/
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./yaba

# Open port
EXPOSE 80

# Start server
CMD ["./yaba"]
