# hashi-helper

`hashi-helper` is a tool mean to enable Disaster Recovery and Configuration Management for Consul and Vault clusters, by exposing configuration via a simple to use and share hcl format.

## Requirements

- Go 1.8

## Building

To build a binary, run the following

```shell
# get this repo
go get github.com/seatgeek/hashi-helper

# go to the repo directory
cd $GOPATH/src/github.com/seatgeek/hashi-helper

# build the `hashi-helper` binary
make build
```

This will create a `hashi-helper` binary in your `$GOPATH/bin` directory.

## Configuration

The following environment variables are required for setting configuration and keys in Consul and Vault.

- `VAULT_TOKEN` environment variable (preferable a root/admin token)
- `VAULT_ADDR` environment variable (example: `http://127.0.0.1:8200`)
- `CONSUL_ADDR_HTTP` environment variable (example: `http://127.0.0.1:8500`)

## Usage

```shell
hashi-helper [--global-flags] command [--command-flags]
```

### Global Flags

- `--concurrency` / `CONCURRENCY`: How many parallel requests to run in parallel against remote servers (optional, default: `2 * CPU Cores`)
- `--log-level` / `LOG_LEVEL`: Debug level of `debug`, `info`, `warn/warning`, `error`, `fatal`, `panic` (optional, default: `info`)
- `--config-dir` / `CONFIG_DIR`: A directory to recursively scan for `hcl` configuration files (optional; default: `./conf.d`)
- `--config-file` / `CONFIG_FILE`: A single `hcl` configuration file to parse instead of a directory (optional; default: `<empty>`)
- `--environment` / `ENVIRONMENT`: The environment to process for (optional; default: `all`)
- `--application` / `APPLICATION`: The application to process for (optional; default: `all`)

### Global Commands

#### `push-all`

Push all Consul and Vault data to remote servers (same as running `vault-push-all` and `consul-push-all`)

### Consul

#### `consul-push-all`

Push all local consul state to remote consul cluster.

#### `consul-push-services`

Push all `service{}` stanza to remote Consul cluster

### vault commands

#### `vault-create-token`

Create a Vault token, optionally encrypt it using keybase

- `--keybase` optional - can be repeated for multiple recipients, token will be encrypted for all recipients to decrypt. If omitted, token is shown in cleartext in the shell
- `--id` optional - The ID of the client token. Can only be specified by a root token. Otherwise, the token ID is a randomly generated UUID.
- `--display-name` optional The display name of the token. Defaults to "token".
- `--ttl` optional - The TTL period of the token, provided as "1h", where hour is the largest suffix. If not provided, the token is valid for the default lease TTL, or indefinitely if the root policy is used.
- `--period` optional - If specified, the token will be periodic; it will have no maximum TTL (unless an "explicit-max-ttl" is also set) but every renewal will use the given period. Requires a root/sudo token to use.
- `--orphan` Will create the token as orphan
- `--policy` optional - Can be repeated for each policy needed. A list of policies for the token. This must be a subset of the policies belonging to the token making the request, unless root. If not specified, defaults to all the policies of the calling token.

#### `vault-find-token`

Scan all tokens in the Vault server, optionally tokens matching certain conditions

Filter flags:

`--filter-name jose` will only match tokens where display name contains `jose`

`--filter-policy root` will only match tokens that have the policy `root`

`--filter-orphan` will only match tokens that are orphaned

Action flags:

`--delete-matches` will match all tokens matching the filter flags. You will be asked to verify each token before deleting it.

#### `vault-list-secrets`

Print a list of local from `conf.d/` (default) or remote secrets from Vault (`--remote`).

Add `--detailed` / `DETAILED` to show secret data rather than just the key names.

#### `vault-profile-edit`

Decrypt (or create), open and encrypt the secure `VAULT_PROFILE_FILE` (`~/.vault_profiles.pgp`) file containing your vault clusters

File format is as described below, a simple yaml file

```yml
---
name_1:
  server: http://<ip>:8200
  token: <your token>
name_2:
  server: http://<ip>:8200
  token: <your token>
```

#### `vault-profile-use`

Decrypt the `VAULT_PROFILE_FILE` and output bash/zsh compatible commands to set `VAULT_ADDR`, `VAULT_TOKEN`, `ENVIRONMENT` based on the profile you selected.

Example: `$(hashi-helper vault-profile-use name_1)`

#### `vault-pull-secrets`

NOT IMPLEMENTED YET

Write remote Vault secrets to local disk in `conf.d/`

#### `vault-push-all`

Pushes all  `mounts`, `policies` and `secrets` to a remote vault server

#### `vault-push-auth`

Write Vault `auth {}` stanza found in `conf.d/` to remote vault server

#### `vault-push-mounts`

Mount and configure `mount {}` stanza found in `conf.d/` to remote vault server

#### `vault-push-policies`

Write Vault `policy {}` stanza found in `conf.d/` to remote vault server

#### `vault-push-secrets`

Write local secrets to remote Vault instance

#### `vault-unseal-keybase`

Unseal Vault using the raw unseal key from [keybase / gpg init/rekey](https://www.vaultproject.io/docs/concepts/pgp-gpg-keybase.html) .

The command expect the raw base64encoded unseal key as env `VAULT_UNSEAL_KEY` or `--unseal-key CLI argument`

`VAULT_UNSEAL_KEY=$token hashi-helper vault-unseal-keybase`
or
`hashi-helper vault-unseal-keybase --unseal-key=$key`

It basically automates `echo "$VAULT_UNSEAL_KEY" | base64 -D | keybase pgp decrypt | xargs vault unseal -address http://<IP>:8200`

## Workflow

The following is a sample workflow that may be used for organizations with Consul and Vault clusters in different environments. If your setup deviates from said description, feel free to modify your workflow.

### Directory Structure

The directory structure is laid out like described below:

- `/${env}/apps/${app}.hcl` (encrypted) Vault secrets for an application in a specific environment.
- `/${env}/auth/${name}.hcl` (encrypted) [Vault auth backends](https://www.vaultproject.io/docs/auth/index.html) for an specific environment `${env}`.
- `/${env}/consul_services/${type}.hcl` (cleartext) List of static Consul services that should be made available in an specific environment `${env}`.
- `/${env}/databases/${name}/_mount.hcl` (encrypted) [Vault secret backend](https://www.vaultproject.io/docs/secrets/index.html) configuration for an specific mount `${name}` in `${env}`.
- `/${env}/databases/${name}/*.hcl` (cleartext) [Vault secret backend](https://www.vaultproject.io/docs/secrets/index.html) configuration for an specific Vault role belonging to mount `${name}` in `${env}`.

### Configuration Examples

The following example assumes:

- A service called api-admin in a `production` environment
- IAM-based authentication to vault
- An elasticache instance called `shared`
- A mysql instance called `db-api` that should provide `read-only` access

Some string will need replacement

### Vault app secret

The following can be stored in an encrypted file at `production/apps/api-admin.hcl`.

```hcl
# environment name must match the directory name
environment "production" {
  
  # application name must match the file name
  application "api-admin" {
  
    # Vault policy granting any user with policy api-admin-read-only read+list access to all secrets
    policy "api-admin-read-only" {
      path "secret/api-admin/*" {
        capabilities = ["read", "list"]
      }
    }

    # an sample secret, will be written to secrets/api-admin/API_URL in Vault
    secret "API_URL" {
      value = "http://localhost:8181"
    }
  }
}
```

### Vault auth 


The following can be stored in an encrypted file at `production/auth/aws-ec2.hcl`.

```hcl
# environment name must match the directory name
environment "production" {

  # The auth name can be anything
  auth "aws-ec2" {

    # Type must match the Vault auth types
    type = "aws-ec2"

    # Client Configuration for the autb backend
    config "client" {
      access_key = "XXXX"
      secret_key = "YYYY"
      max_ttl    = "1d"
    }

    # Sample auth role 
    role "api-admin-prod" {
      policies                       = "global,sample-policy"
      max_ttl                        = "1h"
      allow_instance_migration       = false
      bound_vpc_id                   = "vpc-XXXXXX"
      bound_iam_instance_profile_arn = "arn:aws:iam::XXXXXXX:instance-profile/XXXX"
    }
  }
}
```

### Consul services

The following can be stored in a cleartext file at `production/consul_services/cache.hcl`

```hcl
# environment name must match the directory name
environment "production" {

  # service name
  service "cache-shared" {
    # ID must be unique 
    id      = "cache-shared"

    # Pseudo node name to attach the service to (we use "cache" or "rds" depending on service type)
    node    = "cache"

    # The IP or domain the service should resolve to
    address = "cache-shared.ang13m.YYYY.use1.cache.amazonaws.com"

    # The port the service exposes, used for SRV records and Consul API / Fabio
    port    = 6379

    # Optional list of tags
    tags    = ["master", "replica"]
  }
}
```

### Vault mount

The following can be stored in an encrypted file at `production/databases/api/_mount.hcl`

```hcl
# environment name must match the directory name
environment "production" {

  # the name / path the mount will be mounted at, must be unique for the environment
  mount "db-api" {
    
    # the mount backend type to use (see Vault docs)
    type = "database"

    # mount configuration, see Vault docs for details
    config "default" {
      plugin_name    = "mysql-rds-database-plugin"
      connection_url = "xxx:yyy@tcp(zzzz:3306)/"
      allowed_roles  = "*"
    }
  }
}
```

### Vault mount role

The following can be stored in a cleartext file at `production/databases/api/read-only.hcl`

```hcl
# environment name must match the directory name
environment "production" {

  # the name *must* match the name from _mount.hcl ! 
  mount "db-api" {
    
    # the role name and configuration
    role "read-only" {
      # by convention, db_name matches the config{} stanza from the _mount example
      db_name     = "default"

      # How long time a token may be alive without being renwed
      default_ttl = "24h"

      # How long time a token can life, disregarding renewal timeout
      max_ttl     = "24h"

      # The SQL to execute when creating a user
      # 
      # '{{name}}', '{{password}}' and '{{expiry}}' (used in postgres)
      creation_statements = <<-SQL
      CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';
      GRANT SELECT ON api.* TO '{{name}}'@'%';
      SQL
    }
  }

  # policy name, granting users with policy "db-api-read-only" access to create credentials from the Vault mount
  # by convention the name is always ${mount_name}-${role_name}
  policy "db-api-read-only" {
    
    # the path to allow Vault read from, always ${mount_name}/creds/${role_name}
    path "db-api/creds/read-only" {
      capabilities = ["read"]
    }
  }

  # This will configure the GitHub team "rds-production-api-read-only" to have the policy "db-api-read-only" 
  # when they "vault auth"
  secret "/auth/github/map/teams/rds-production-api-read-only" {
    value = "db-api-read-only"
  }
}
```
