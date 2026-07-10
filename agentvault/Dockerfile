FROM golang:1.25 AS builder

WORKDIR /src

COPY go.mod ./
COPY agentvault/main.go ./agentvault-bootstrap-job/main.go
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/agentvault-bootstrap-job ./agentvault-bootstrap-job

FROM infisical/agent-vault:latest

COPY --from=builder /out/agentvault-bootstrap-job /usr/local/bin/agentvault-bootstrap-job

ENTRYPOINT ["/usr/local/bin/agentvault-bootstrap-job"]
