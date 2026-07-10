# Infisical Bootstrap Job

This folder contains the Infisical bootstrap job code used by this repo.

It supports two entrypoints:

- `infisical-bootstrap-job`: the reusable standalone bootstrap job
- `infisical-compose-bootstrap`: the compose-specific bootstrap flow that prepares Infisical for Agent Vault e2e automation

## Intent

- bootstrap a fresh Infisical instance in platform mode
- create or reuse an Infisical project
- create or reuse an environment
- create or reuse an Infisical identity
- grant that identity access to the project
- optionally enable Kubernetes Auth for that identity
- optionally write bootstrap results back into Kubernetes secrets
- optionally emit compose metadata for downstream Agent Vault automation

## Files

- `main.go`: standalone bootstrap entrypoint
- `compose/main.go`: compose-focused entrypoint for this repo's e2e stack
- `Dockerfile`: image build for both binaries
- `example-project-bootstrap-job.yaml`: Kubernetes example for project bootstrap
- `example-platform-bootstrap-job.yaml`: Kubernetes example for platform bootstrap

## Build

```bash
docker build -t your-registry/infisical-bootstrap-job:latest -f infisical/Dockerfile .
```

The default image entrypoint is `infisical-bootstrap-job`.

For the compose automation in this repo, `docker-compose.yml` overrides the entrypoint to `infisical-compose-bootstrap`.

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
- `ORGANIZATION_ID` or `ORGANIZATION_ID_SOURCE_NAMESPACE` + `ORGANIZATION_ID_SOURCE_CONFIGMAP`
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
- `ORGANIZATION_ID_SOURCE_KEY`
- `SMOKE_TEST_SECRET_KEY`
- `SMOKE_TEST_SECRET_VALUE`
- `OUTPUT_STATIC_SECRET_NAME`
- `OUTPUT_STATIC_SECRET_NAMESPACE`
- `OUTPUT_STATIC_SECRET_AUTH_REF_NAME`
- `OUTPUT_STATIC_SECRET_AUTH_REF_NAMESPACE`
- `OUTPUT_STATIC_SECRET_TARGET_SECRET_NAME`
- `SECRETS_JSON`

### Project mode

Required environment variables:

- `BOOTSTRAP_MODE=project`
- `INFISICAL_URL`
- `INFISICAL_EMAIL`
- `INFISICAL_PASSWORD`
- `ORGANIZATION_ID` or `ORGANIZATION_ID_SOURCE_NAMESPACE` + `ORGANIZATION_ID_SOURCE_CONFIGMAP`
- `PROJECT_NAME`
- `PROJECT_SLUG`
- `ENVIRONMENT_NAME`
- `ENVIRONMENT_SLUG`
- `IDENTITY_NAME`

Optional environment variables:

- `ORGANIZATION_ID_SOURCE_KEY`
- `IDENTITY_ROLE`
- `ENABLE_KUBERNETES_AUTH`
- `ENABLE_UNIVERSAL_AUTH`
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
- `OUTPUT_STATIC_SECRET_NAME`
- `OUTPUT_STATIC_SECRET_NAMESPACE`
- `OUTPUT_STATIC_SECRET_AUTH_REF_NAME`
- `OUTPUT_STATIC_SECRET_AUTH_REF_NAMESPACE`
- `OUTPUT_STATIC_SECRET_TARGET_SECRET_NAME`
- `SECRETS_JSON`

When `ENABLE_UNIVERSAL_AUTH=true`, the job also attaches Universal Auth to the machine identity, creates a fresh client secret, and prints `universalAuthClientId` plus `universalAuthClientSecret` in the JSON result. This is useful for wiring Agent Vault to an Infisical-backed vault in automation.

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
- when `ORGANIZATION_ID` is omitted, the job can hydrate it from a prior bootstrap status `ConfigMap` using `ORGANIZATION_ID_SOURCE_NAMESPACE`, `ORGANIZATION_ID_SOURCE_CONFIGMAP`, and optional `ORGANIZATION_ID_SOURCE_KEY`
- when `OUTPUT_STATIC_SECRET_NAME` is set, the job also creates or updates an `InfisicalStaticSecret` that points at the bootstrapped project ID
- `operator` mode is the only mode that reconciles organization-level machine identity role
- `project` mode reuses an existing machine identity name and only manages project-level membership
- this image does not assume any specific app; app-specific manifests can supply different env vars
- the compose-only metadata handoff used by this repo lives in `internal/jobhandoff`, not in the generic bootstrap package
