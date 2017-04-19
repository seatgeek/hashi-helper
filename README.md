# hashi-helper

## general

```shell
hashi-helper [--global-flags] command [--command-flags]
```

## Expectations

`VAULT_TOKEN` environment variable (preferable a root/admin token)

`VAULT_ADDR` environment variable (example: `http://127.0.0.1:8200`)

### global flags

`--concurrency` / `CONCURRENCY`: How many parallel requests to run in parallel against remote servers (2 * CPU Cores)

`--log-level` / `LOG_LEVEL`: Debug level (debug, info, warn/warning, error, fatal, panic)

`--config-dir` / `CONFIG_DIR`: The conf.d directory to read/write to (default: `./conf.d`)

`--environment` / `ENVIRONMENT`: The environment to process for (default: all)

`--application` / `APPLICATION`: The application to process for (default: all)

## vault

`vault-list-secrets`: Print a list of local from `conf.d/` (default) or remote secrets from Vault (`--remote`). Add `--detailed` / `DETAILED` to show secret data rather than just the key names.

`vault-pull-secrets`: Write remote Vault secrets to local disk in `conf.d/`

`vault-push-secrets`: Write local secrets to remote Vault instance

`vault-push-policies`: Write a Vault read-only policy for each env + application combo that exist in `conf.d/`
