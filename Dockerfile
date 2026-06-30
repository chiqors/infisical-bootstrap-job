FROM golang:latest AS builder

WORKDIR /src

COPY go.mod .
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/infisical-bootstrap-job ./cmd/infisical-bootstrap-job

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/infisical-bootstrap-job /usr/local/bin/infisical-bootstrap-job

ENTRYPOINT ["/usr/local/bin/infisical-bootstrap-job"]
