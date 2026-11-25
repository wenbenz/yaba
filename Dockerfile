# Build stages run natively on the build host ($BUILDPLATFORM). The UI output is
# architecture-independent and Go cross-compiles, so multi-arch builds need no emulation.
FROM --platform=$BUILDPLATFORM node:lts AS ui-builder
WORKDIR /ui
# node:lts no longer bundles yarn
RUN npm install -g --force yarn@1.22.22
COPY ui/package.json ui/yarn.lock ./
RUN yarn install --frozen-lockfile
COPY ui/ ./
RUN yarn build

FROM --platform=$BUILDPLATFORM golang:1.26 AS go-builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY errors ./errors/
COPY gen ./gen/
COPY config/*.go ./config/
COPY graph/ ./graph/
COPY internal ./internal/
COPY ui/*.go ./ui/
COPY --from=ui-builder /ui/dist ./ui/dist/
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/yaba

FROM gcr.io/distroless/static-debian12
WORKDIR /yaba
COPY migrations ./migrations/
COPY --from=go-builder /out/yaba ./yaba
EXPOSE 80
CMD ["./yaba"]
