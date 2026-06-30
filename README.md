# Infisical Bootstrap Job

This folder contains standalone source code for an Infisical bootstrap job image.

Intent:

- bootstrap a fresh Infisical instance in platform mode
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
- `example-project-bootstrap-job.yaml`: working example for project bootstrap with the current image
- `example-platform-bootstrap-job.yaml`: working example for platform bootstrap with the current image

## Build

```bash
docker build -t your-registry/infisical-bootstrap-job:latest jobs/infisical-bootstrap-job
```

## Runtime inputs

Set `BOOTSTRAP_MODE` to choose the flow:

- `platform`: bootstrap a fresh Infisical instance and write the returned bootstrap token to Kubernetes if desired
- `operator`: bootstrap the shared organization machine identity, reconcile its org role, bind it into the operator project, enable Kubernetes auth, and write Kubernetes output secrets
- `project`: bootstrap a project, environment, and project membership for a reused machine identity, plus optional app secrets

If `BOOTSTRAP_MODE` is omitted, it defaults to `project`.

### Platform mode

Required environment variables:

- `BOOTSTRAP_MODE=platform`
- `INFISICAL_URL`
- `BOOTSTRAP_EMAIL`
- `BOOTSTRAP_PASSWORD`
- `ORGANIZATION_NAME`

Optional environment variables:

- `IGNORE_IF_BOOTSTRAPPED`
- `WRITE_KUBERNETES_SECRET`
- `OUTPUT_SECRET_NAMESPACE`
- `OUTPUT_SECRET_NAME`
- `OUTPUT_SECRET_KEY`

### Operator mode

Required environment variables:

- `BOOTSTRAP_MODE=operator`
- `INFISICAL_URL`
- `BOOTSTRAP_SECRET_NAMESPACE` or `INFISICAL_EMAIL`
- `BOOTSTRAP_SECRET_NAME` or `INFISICAL_PASSWORD`
- `ORGANIZATION_ID`
- `PROJECT_NAME`
- `PROJECT_SLUG`
- `ENVIRONMENT_NAME`
- `ENVIRONMENT_SLUG`
- `IDENTITY_NAME`
- `ORGANIZATION_IDENTITY_ROLE`
- `IDENTITY_ROLE`
- `KUBERNETES_AUTH_HOST`
- `ALLOWED_NAMESPACES`
- `ALLOWED_SERVICE_ACCOUNTS`
- `OUTPUT_SECRET_NAMESPACE`
- `OUTPUT_SECRET_NAME`
- `OUTPUT_SECRET_KEY`
- `OUTPUT_PROJECT_SECRET_NAME`
- `OUTPUT_PROJECT_SECRET_KEY`

Optional environment variables:

- `OUTPUT_STATUS_CONFIGMAP`
- `SMOKE_TEST_SECRET_KEY`
- `SMOKE_TEST_SECRET_VALUE`
- `SECRETS_JSON`

### Project mode

Required environment variables:

- `BOOTSTRAP_MODE=project`
- `INFISICAL_URL`
- `INFISICAL_EMAIL`
- `INFISICAL_PASSWORD`
- `ORGANIZATION_ID`
- `PROJECT_NAME`
- `PROJECT_SLUG`
- `ENVIRONMENT_NAME`
- `ENVIRONMENT_SLUG`
- `IDENTITY_NAME`

Optional environment variables:

- `IDENTITY_ROLE`
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
- `SECRETS_JSON`

`SECRETS_JSON` must be a path-aware JSON array:

```json
[
  {
    "key": "INFISICAL_OPERATOR_TEST",
    "value": "ok"
  },
  {
    "key": "test-file-in-folder",
    "value": "hello",
    "path": "/test-folder"
  }
]
```

`path` is optional. If omitted or empty, it defaults to `/`. Writing a secret to a non-root `path` is what makes that folder-style path appear in the Infisical UI.

## Notes

- when `WRITE_KUBERNETES_SECRET=true`, the container expects to run inside Kubernetes with a mounted service account token
- when `ENABLE_KUBERNETES_AUTH=true`, the same in-cluster service account CA bundle is used as the Kubernetes Auth CA certificate
- `operator` mode is the only mode that reconciles organization-level machine identity role
- `project` mode reuses an existing machine identity name and only manages project-level membership
- this image does not assume any specific app; app-specific manifests can supply different env vars
- that platform example intentionally mirrors the current repo manifest at `gitops/apps/infisical/manifests/infisical-bootstrap-job.yaml`
