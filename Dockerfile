FROM golang:1.16-alpine as builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o build/ ./...


FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/build/k8s-scheduler .
USER nonroot:nonroot

ENTRYPOINT ["/k8s-scheduler"]
