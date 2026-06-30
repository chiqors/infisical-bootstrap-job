# Infisical Bootstrap Job

This folder contains standalone source code for an Infisical bootstrap job image.

Intent:

- create or reuse an Infisical project
- create or reuse an environment
- create or reuse an Infisical identity
- grant that identity access to the project
- optionally enable Kubernetes Auth for that identity
- optionally write bootstrap results back into Kubernetes secrets

This is intentionally separate from GitOps manifests so the image can be built and reused by different jobs.

Implementation note:

- the job is written in Go
- the Dockerfile uses the latest `golang` image in a multi-stage build

## Files

- `cmd/infisical-bootstrap-job/main.go`: bootstrap entrypoint
- `go.mod`: Go module definition
- `Dockerfile`: build recipe for the job image
- `example-infisical-bootstrap-job.yaml`: example Kubernetes Job manifest using a published image

## Build

```bash
docker build -t your-registry/infisical-bootstrap-job:latest jobs/infisical-bootstrap-job
```

## Runtime inputs

Required environment variables:

- `INFISICAL_URL`
- `INFISICAL_EMAIL`
- `INFISICAL_PASSWORD`
- `ORGANIZATION_ID`
- `PROJECT_NAME`
- `PROJECT_SLUG`
- `ENVIRONMENT_NAME`
- `ENVIRONMENT_SLUG`
- `IDENTITY_NAME`
- `IDENTITY_ROLE`

Optional environment variables:

- `ENABLE_KUBERNETES_AUTH`
- `KUBERNETES_AUTH_HOST`
- `ALLOWED_NAMESPACES`
- `ALLOWED_SERVICE_ACCOUNTS`
- `WRITE_KUBERNETES_SECRET`
- `OUTPUT_SECRET_NAMESPACE`
- `OUTPUT_SECRET_NAME`
- `OUTPUT_SECRET_KEY`
- `OUTPUT_PROJECT_SECRET_NAME`
- `OUTPUT_PROJECT_SECRET_KEY`
- `SMOKE_TEST_SECRET_KEY`
- `SMOKE_TEST_SECRET_VALUE`

## Notes

- when `WRITE_KUBERNETES_SECRET=true`, the container expects to run inside Kubernetes with a mounted service account token
- when `ENABLE_KUBERNETES_AUTH=true`, the same in-cluster service account CA bundle is used as the Kubernetes Auth CA certificate
- this image does not assume any specific app; app-specific manifests can supply different env vars
