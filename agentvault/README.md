# Agent Vault Bootstrap Job

This folder contains the Agent Vault automation used by this repository.

It has two entrypoints:

- `agentvault-bootstrap-job`: bootstraps Agent Vault resources after the server is running
- `agent-vault-server-launcher`: waits for Infisical bootstrap metadata, exports the required env vars, and then starts `agent-vault server`

## Intent

This automation exists to remove manual first-boot steps from Agent Vault e2e flows.

In particular, it can automate:

- initial owner registration
- owner login
- creation of an Infisical-backed vault
- credential-store sync
- service registration
- agent token creation or rotation
- writing generated metadata for downstream tests

## Files

- `main.go`: Agent Vault bootstrap job entrypoint
- `launcher/main.go`: startup wrapper for `agent-vault server`
- `Dockerfile`: image build for the bootstrap job

## Build

Build the bootstrap image:

```bash
docker build -t your-registry/agentvault-bootstrap-job:latest -f agentvault/Dockerfile .
```

The default image entrypoint is `agentvault-bootstrap-job`.

## Runtime Model

The bootstrap job assumes:

- Agent Vault server is already installed in the image
- Agent Vault server is reachable
- Infisical bootstrap has already produced metadata that includes:
  - Infisical URL
  - Infisical project ID
  - Infisical environment slug
  - Universal Auth client credentials

The launcher exists because Agent Vault server may need those Infisical credentials available as environment variables before it starts.

## Required Environment Variables

The bootstrap job requires:

- `BOOTSTRAP_METADATA_PATH`
- `AGENT_VAULT_OWNER_EMAIL`
- `AGENT_VAULT_OWNER_PASSWORD`
- `AGENT_VAULT_SYNC_VAULT_NAME`

It also supports:

- `AGENT_VAULT_URL`

If `AGENT_VAULT_URL` is omitted, it defaults to:

```bash
http://agent-vault:14321
```

## Proxy Service Bootstrap Variables

Optional variables for direct bearer injection service bootstrap:

- `AGENT_VAULT_PROXY_SERVICE_NAME`
- `AGENT_VAULT_PROXY_SERVICE_HOST`
- `AGENT_VAULT_PROXY_TOKEN_KEY`
- `AGENT_VAULT_PROXY_AGENT_NAME`
- `AGENT_VAULT_PROXY_OUTPUT_PATH`

Defaults:

- service name: `echo-api`
- service host: `echo-api.local:8080`
- token key: `E2E_API_TOKEN`
- agent name: `e2e-proxy-agent`
- output path: `/data/agent-vault-proxy.json`

## Passthrough Service Bootstrap Variables

Optional variables for tutorial-style passthrough substitution bootstrap:

- `AGENT_VAULT_PASSTHROUGH_SERVICE_NAME`
- `AGENT_VAULT_PASSTHROUGH_SERVICE_HOST`
- `AGENT_VAULT_PASSTHROUGH_TOKEN_KEY`
- `AGENT_VAULT_PASSTHROUGH_PLACEHOLDER`
- `AGENT_VAULT_PASSTHROUGH_AGENT_NAME`
- `AGENT_VAULT_PASSTHROUGH_OUTPUT_PATH`

Defaults:

- service name: `echo-passthrough`
- service host: `echo-passthrough.local:8080`
- token key: `E2E_API_TOKEN`
- placeholder: `__e2e_api_token__`
- agent name: `e2e-passthrough-agent`
- output path: `/data/agent-vault-passthrough.json`

## What The Bootstrap Job Does

At a high level, `main.go` performs this sequence:

1. Read shared bootstrap metadata from `BOOTSTRAP_METADATA_PATH`.
2. Wait for Agent Vault health.
3. Try to log in as the configured owner.
4. If login fails, register the owner and then log in.
5. Create the target vault if it does not already exist.
6. Configure that vault to use Infisical as its credential store.
7. Sync the credential store.
8. Create or reconcile:
   - a direct-injection proxy service
   - a passthrough substitution proxy service
9. Create or rotate agent tokens for those services.
10. Write test metadata JSON files for downstream consumers.

## Output Files

The direct injection metadata file contains fields such as:

- agent token
- vault name
- target URL
- sample proxy curl command

Default path:

```bash
/data/agent-vault-proxy.json
```

The passthrough metadata file additionally includes the placeholder configuration.

Default path:

```bash
/data/agent-vault-passthrough.json
```

## Direct Injection vs Passthrough

This job currently configures two useful service patterns.

Direct injection:

- the caller authenticates to Agent Vault
- the caller does not send the real secret
- Agent Vault injects the bearer token from the synced Infisical secret

Passthrough substitution:

- the caller authenticates to Agent Vault
- the caller sends a placeholder like `Authorization: Bearer __e2e_api_token__`
- Agent Vault replaces the placeholder with the real synced secret before forwarding

The passthrough model is closer to the public Agent Vault tutorial flow, while direct injection is often cleaner for tightly controlled agent systems.

## Agent Vault Launcher

The launcher in `launcher/main.go` is intentionally small.

It:

- waits for `BOOTSTRAP_METADATA_PATH`
- reads Infisical connection and Universal Auth credentials
- exports:
  - `INFISICAL_URL`
  - `INFISICAL_UNIVERSAL_AUTH_CLIENT_ID`
  - `INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET`
- replaces itself with `agent-vault server --host 0.0.0.0`

This avoids race conditions where Agent Vault starts before its Infisical auth inputs are ready.

## Notes

- this automation is optimized for bootstrap and e2e flows
- it is intentionally separate from the generic Infisical bootstrap logic
- owner creation happens only when login fails, which makes reruns friendlier for existing environments
- agent tokens are rotated on rerun when the target agent already exists
