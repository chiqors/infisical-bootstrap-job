FROM golang:1.25 AS builder

WORKDIR /src

COPY infisical-bootstrap-job/go.mod ./
COPY infisical-bootstrap-job/infisical/main.go ./infisical-bootstrap-job/main.go
COPY infisical-bootstrap-job/infisical/compose/main.go ./infisical-compose-bootstrap/main.go
COPY infisical-bootstrap-job/internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/infisical-bootstrap-job ./infisical-bootstrap-job && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/infisical-compose-bootstrap ./infisical-compose-bootstrap

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/infisical-bootstrap-job /usr/local/bin/infisical-bootstrap-job
COPY --from=builder /out/infisical-compose-bootstrap /usr/local/bin/infisical-compose-bootstrap

ENTRYPOINT ["/usr/local/bin/infisical-bootstrap-job"]
