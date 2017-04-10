# hashi-helper

## general

```shell
hashi-helper [--global-flags] command [--command-flags]
```

### global flags

`--concurrency` / `CONCURRENCY`: How many parallel requests to run in parallel against remote servers (2 * CPU Cores)

`--log-level` / `LOG_LEVEL`: Debug level (debug, info, warn/warning, error, fatal, panic)

`--environment` / `ENVIRONMENT`: The environment to process for (default: all)

`--config-dir` / `CONFIG_DIR`: The conf.d directory to read/write to (default: `./conf.d`)

## vault

`vault-local-list`: Print a list of local secrets from `conf.d/`

`vault-local-write`: Write remote Vault secrets to local disk in `conf.d/`

`vault-remote-list`: Print a list of remote secrets. Add `--detailed` / `DETAILED` to show the secret data, other than just a list of keys

`vault-remote-write`: Write local secrets to remote Vault instance

`vault-remote-clean`: Delete remote Vault secrets not in the local catalog
