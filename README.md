# infisical-bootstrap-job

Reusable Go bootstrap jobs for:

- Infisical
- Agent Vault

This repository is intended for automation scenarios where you want to stand up Infisical and then bootstrap the downstream resources needed by Agent Vault or other systems.

## What Is In This Repo

This repo contains two job families:

- `infisical/`: bootstraps Infisical itself
- `agentvault/`: bootstraps Agent Vault against an already running Agent Vault server

It also contains shared packages used by both:

- `internal/bootstrap`: reusable Infisical bootstrap logic
- `internal/jobhandoff`: small metadata and env helpers used when one bootstrap step hands data to another

## Repository Layout

```text
.
├── agentvault/
│   ├── main.go
│   ├── Dockerfile
│   └── launcher/
├── infisical/
│   ├── main.go
│   ├── compose/
│   ├── Dockerfile
│   ├── README.md
│   └── example-*.yaml
├── internal/
│   ├── bootstrap/
│   └── jobhandoff/
└── go.mod
```

## Infisical Bootstrap Job

The `infisical` job is the main reusable bootstrap entrypoint.

It supports these use cases:

- bootstrap a fresh Infisical instance in platform mode
- create or reuse an organization project
- create or reuse an environment
- create or reuse a machine identity
- grant project membership
- optionally enable Kubernetes Auth
- optionally enable Universal Auth
- optionally seed secrets
- optionally write results back to Kubernetes secrets

For detailed runtime configuration, examples, and supported modes, see [infisical/README.md](./infisical/README.md).

## Agent Vault Bootstrap Job

The `agentvault` job is a separate bootstrap step for Agent Vault.

Its role is different from the Infisical bootstrap job:

- it does not bootstrap Infisical itself
- it assumes Agent Vault server is already running
- it uses Infisical bootstrap outputs to configure Agent Vault resources

In the current implementation it can:

- wait for Agent Vault health
- create the initial owner account if needed
- log in with that owner
- create or reuse an Infisical-backed vault
- trigger credential-store sync
- register proxy services
- create or rotate test agent tokens
- write generated access metadata for downstream tests

This is especially useful for compose-based or CI e2e flows where Agent Vault needs to be fully usable without manual first-login setup.

## Agent Vault Launcher

The `agentvault/launcher` binary is a small helper for environments where Agent Vault server must not start until Infisical bootstrap metadata is available.

It:

- waits for shared metadata to exist
- exports the required Infisical auth environment variables
- `exec`s `agent-vault server`

This keeps Agent Vault startup deterministic in automated environments.

## Build

Build the Infisical image:

```bash
docker build -t your-registry/infisical-bootstrap-job:latest -f infisical/Dockerfile .
```

Build the Agent Vault image:

```bash
docker build -t your-registry/agentvault-bootstrap-job:latest -f agentvault/Dockerfile .
```

## Running

The Infisical job entrypoint is documented in [infisical/README.md](./infisical/README.md).

The Agent Vault job is driven by environment variables. In the current e2e-oriented flow it expects values such as:

- `BOOTSTRAP_METADATA_PATH`
- `AGENT_VAULT_URL`
- `AGENT_VAULT_OWNER_EMAIL`
- `AGENT_VAULT_OWNER_PASSWORD`
- `AGENT_VAULT_SYNC_VAULT_NAME`

Optional service bootstrap variables include:

- `AGENT_VAULT_PROXY_SERVICE_NAME`
- `AGENT_VAULT_PROXY_SERVICE_HOST`
- `AGENT_VAULT_PROXY_TOKEN_KEY`
- `AGENT_VAULT_PROXY_AGENT_NAME`
- `AGENT_VAULT_PROXY_OUTPUT_PATH`
- `AGENT_VAULT_PASSTHROUGH_SERVICE_NAME`
- `AGENT_VAULT_PASSTHROUGH_SERVICE_HOST`
- `AGENT_VAULT_PASSTHROUGH_TOKEN_KEY`
- `AGENT_VAULT_PASSTHROUGH_PLACEHOLDER`
- `AGENT_VAULT_PASSTHROUGH_AGENT_NAME`
- `AGENT_VAULT_PASSTHROUGH_OUTPUT_PATH`

## Design Notes

- `internal/bootstrap` is the generic Infisical bootstrap layer
- `internal/jobhandoff` is intentionally small and focused on passing bootstrap outputs between steps
- the Infisical and Agent Vault jobs are split because they operate at different lifecycle stages
- the launcher exists because some environments need Infisical-derived auth env vars before Agent Vault server can start correctly

## Typical Automation Flow

In a full e2e environment, the flow usually looks like this:

1. Start Infisical and its dependencies.
2. Run the Infisical bootstrap job.
3. Persist the resulting metadata needed by Agent Vault.
4. Start Agent Vault with the launcher if startup depends on that metadata.
5. Run the Agent Vault bootstrap job.
6. Run proxy or integration tests.

## Status

This repo is intended for automation and integration workflows. Some parts are generic and reusable, while some Agent Vault behavior is currently optimized around e2e bootstrap and service-sync scenarios.
