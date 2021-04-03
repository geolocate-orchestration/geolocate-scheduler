FROM golang:1.16-alpine as builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY utils/ utils/
COPY main.go main.go
COPY scheduler/ scheduler/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build


FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/aida-scheduler .
USER nonroot:nonroot

ENTRYPOINT ["/aida-scheduler"]
