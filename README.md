# hashi-helper

## general

```shell
hashi-helper [--global-flags] command [--command-flags]
```

## Expectations

`VAULT_TOKEN` environment variable (preferable a root/admin token)

`VAULT_ADDR` environment variable (example: `http://127.0.0.1:8200`)

### global flags

`--concurrency` / `CONCURRENCY`: How many parallel requests to run in parallel against remote servers (optional, default: `2 * CPU Cores`)

`--log-level` / `LOG_LEVEL`: Debug level of `debug`, `info`, `warn/warning`, `error`, `fatal`, `panic` (optional, default: `info`)

`--config-dir` / `CONFIG_DIR`: The conf.d directory to read/write to (optional; default: `./conf.d`)

`--environment` / `ENVIRONMENT`: The environment to process for (optional; default: `all`)

`--application` / `APPLICATION`: The application to process for (optional; default: `all`)

## vault

### vault-list-secrets

Print a list of local from `conf.d/` (default) or remote secrets from Vault (`--remote`).

Add `--detailed` / `DETAILED` to show secret data rather than just the key names.

### vault-pull-secrets

NOT IMPLEMENTED YET

Write remote Vault secrets to local disk in `conf.d/`

### vault-push-secrets

Write local secrets to remote Vault instance

### vault-push-policies

Write Vault `policy {}` stanza found in `conf.d/` to remote vault server

### vault-push-mounts

Mount and configure `mount {}` stanza found in `conf.d/` to remote vault server

# Install & usage

```
go get -u github.com/kardianos/govendor
govendor sync
blackbox_postdeploy
```

# Example

```hcl
environment "production" {
  # vault policy - will be pushed to vault as `production-global` or `global` depending on your isolation setting
  # everything inside the policy stanza is all the same HCL as an actual Vault Policy file
  # the parsing and validation is vendored directly from Vault itself
  policy "global" {
    path "auth/app-id/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }

    path "auth/app-idx/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }
  }

  # vault mount - will be created as `production-app1` or `app1` depending on your isolation setting
  # the engine is MySQL
  mount "mysql" "app1" {
    config "connection" {
      connection_url = "root:root@tcp(192.168.33.10:3306)/"
    }

    config "lease" {
      lease     = "1h"
      lease_max = "24h"
    }

    role "read-only" {
      sql = "CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}'; GRANT SELECT ON *.* TO '{{name}}'@'%';"
    }
  }

  # vault mount - will be created as `production-app2` or `app2` depending on your isolation setting
  # the engine is postgresql
  mount "postgresql" "app2" {
    config "connection" {
      connection_url = "postgresql://root:vaulttest@vaulttest.ciuvljjni7uo.us-west-1.rds.amazonaws.com:5432/postgres"
    }

    config "lease" {
      lease     = "1h"
      lease_max = "24h"
    }

    role "readonly" {
      sql = "CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";"
    }
  }

  # rabbitmq mount - will be created as `production-derp` or `derp` depending on your isolation setting
  # the engine is rabbitmq
  mount "rabbitmq" "derp" {
    config "connection" {
      connection_uri = "http://localhost:15672"
      username       = "admin"
      password       = "password"
    }

    config "lease" {
      ttl     = 3600
      max_ttl = 86400
    }

    role "readwrite" {
      vhosts = <<-USER_DATA
        {"/":{"write": ".*", "read": ".*"}}
      USER_DATA
    }
  }

  # application specific secrets and policies
  application "grafana" {
    # same as environment -> policy
    # will be created as `production-grafana-derp` or `grafana-derp` depending on isolation setting
    policy "derp" {
      path "auth/app-id/*" {
        capabilities = ["create", "read", "update", "delete", "list"]
      }
    }

    # a vault secret
    # will be created as `secret/production/grafana/api_key` or `secret/grafana/api_key` depending on isolation settig
    secret "api_key" {
      value = "xxxx"
    }
  }
}
```
