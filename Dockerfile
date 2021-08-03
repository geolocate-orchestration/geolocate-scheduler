FROM golang:1.16-alpine as builder

WORKDIR /workspace

ARG git_user
ARG git_token

RUN apk add git && \
    go env -w GOPRIVATE=github.com/geolocate-orchestration && \
    git config --global url."https://${git_user}:${git_token}@github.com".insteadOf "https://github.com"

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o build/ ./...


FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/build/geolocate-scheduler .
USER nonroot:nonroot

ENTRYPOINT ["/geolocate-scheduler"]
