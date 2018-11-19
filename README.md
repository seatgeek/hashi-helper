## hashi-helper

`hashi-helper` is a tool meant to enable Disaster Recovery and Configuration Management for Consul and Vault clusters, by exposing configuration via a simple to use and share hcl format.

- [hashi-helper](#hashi-helper)
- [Requirements](#requirements)
- [Building](#building)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Global Flags](#global-flags)
  - [Global Commands](#global-commands)
    - [`push-all`](#push-all)
  - [Consul](#consul)
    - [`consul-push-all`](#consul-push-all)
    - [`consul-push-services`](#consul-push-services)
    - [`consul-push-kv`](#consul-push-kv)
  - [vault commands](#vault-commands)
    - [`vault-create-token`](#vault-create-token)
    - [`vault-find-token`](#vault-find-token)
    - [`vault-list-secrets`](#vault-list-secrets)
    - [`vault-profile-edit`](#vault-profile-edit)
    - [`vault-profile-use`](#vault-profile-use)
    - [`vault-pull-secrets`](#vault-pull-secrets)
    - [`vault-push-all`](#vault-push-all)
    - [`vault-push-auth`](#vault-push-auth)
    - [`vault-push-mounts`](#vault-push-mounts)
    - [`vault-push-policies`](#vault-push-policies)
    - [`vault-push-secrets`](#vault-push-secrets)
    - [`vault-unseal-keybase`](#vault-unseal-keybase)
      - [Options](#options)
      - [Examples](#examples)
- [Templating](#templating)
- [Workflow](#workflow)
  - [Directory Structure](#directory-structure)
  - [Configuration Examples](#configuration-examples)
  - [Vault app secret](#vault-app-secret)
  - [Consul app KV](#consul-app-kv)
  - [Vault auth](#vault-auth)
  - [Consul services](#consul-services)
  - [Vault mount](#vault-mount)
  - [Vault mount role](#vault-mount-role)

## Requirements

- Go 1.11

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

#### `consul-push-kv`

Push all `kv{}` stanza to remote Consul cluster

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

`--filter-path auth/github/login` will only match tokens that have the path `auth/github/login`

`--filter-meta-username jippi` will only match tokens that have the `meta[username]` value `jippi` (GitHub auth backend injects this, as an example)

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
# Sample config (yaml)
#
# all keys are optional
#

profile_name_1:
  server: http://active.vault.service.consul:8200
  consul_server: http://consul.service.consul:8500
  token: <your vault token>
  unseal_token: <your unseal token>

profile_name_2:
  server: http://active.vault.service.consul:8200
  consul_server: http://consul.service.consul:8500
  token: <your vault token>
  unseal_token: <your unseal token>
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

It basically automates `echo "$VAULT_UNSEAL_KEY" | base64 -D | keybase pgp decrypt | xargs vault unseal -address http://<IP>:8200`

##### Options

- `--unseal-key`/ `VAULT_UNSEAL_KEY` (default: `<empty>`) The raw base64encoded unseal key as env or CLI argument
- `--consul-service-name` / `CONSUL_SERVICE_NAME` (default: `<empty>`) If specified, the tool will try to lookup all vault servers in the configured Consul catalog and unseal all of them. This is the service name, e.g. `vault`
- `--consul-service-tag` / `CONSUL_SERVICE_TAG` (default: `standby`) The Consul catalog tag to filter vault instances from if `CONSUL_SERVICE_NAME` is used.
- `--vault-protocol` / `VAULT_PROTOCOL` (default: `http`) The protocol to use when constructing the `VAULT_ADDR` value when using `CONSUL_SERVICE_NAME` unseal strategy.

All `VAULT_*` env keys are preserved when using `CONSUL_SERVICE_TAG`, `address` is the only field being overwritten per vault instance found in the catalog. So you can still configure TLS and other Vault changes as usual with the environment.

##### Examples

- `VAULT_UNSEAL_KEY=$token hashi-helper vault-unseal-keybase`
- `hashi-helper vault-unseal-keybase --unseal-key=$key`
- `VAULT_CONSUL_SERVICE=vault VAULT_UNSEAL_KEY=$token hashi-helper vault-unseal-keybase`
- `hashi-helper vault-unseal-keybase --unseal-key=$key --consul-service-name vault`

## Templating

`hashi-helper` version >= 2.0 support templating for configuration files. All configuration files loaded are automatically rendered as a template using [go text/template](https://golang.org/pkg/text/template/).

Templates use `[[ ]]` brackes instead of the default `{{ }}` style to avoid clashes with HCL `{ }` stanza definitions.

### Variables

You can provide variables for templating through the CLI in various ways. All of these options can be provided any number of times

`--variable-file <file>` or `--var-file <file>`

`--variable key=value` or `--var key=value`

The tool can load files with extensions `.hcl`, `.yaml`/`.yml` and `.json`. Variables are loaded in the order they are provided in CLI, so it's possible to cascade / overwrite configuration files by using a specific loading order. `--variable-file <file>` are loaded **before** CLI `--variable key=value` arguments.

#### HCL variable file

```hcl
consul_domain    = "consul"
environment_name = "staging"
environment_tld  = "stag"
db_default_ttl   = "9h"
db_max_ttl       = "72h"

# a list of things

stuff = ["a", "b", "c"]

here_doc = <<-DOC
  something multiline
  that will be available
  as a single string
  DOC
```

#### YAML variable file

```yaml
---
environment_name: "staging"
environment_tld: "stag"
db_default_ttl: "9h"
db_max_ttl: "72h"

# a list of things

stuff:
  - "a"
  - "b"
  - "c"

here_doc: |
  something multiline
  that will be available
  as a single string
```

#### JSON variable file

```json
---
{
  "db_default_ttl": "9h",
  "db_max_ttl": "72h",
  "environment_name": "staging",
  "environment_tld": "stag",
  "here_doc": "something multiline\nthat will be available\nas a single string",
  "stuff": [
    "a",
    "b",
    "c"
  ]
}
```

### Functions

#### lookup

`lookup` is used to lookup template variables inside a template. If the key do not exist, the template rendering will fail.

`[[ lookup "my-key" ]]` will output `hello-world`

#### lookup_default

`lookup_default` is used to lookup template variables inside a template. If the key do not exist, the `default` (2nd argument) will be returned.

`[[ lookup_default "my-key" "something" ]]` will output `hello-world`

#### service

`service` is used to construct Consul service names programatically.

By default `consul` is used as the domain, but can be overwritten with a variable `consul_domain`

`[[ service "test" ]]` will output `test.service.consul`

Given `--variable consul_domain=test.consul`

`[[ service "test" ]]` will output `test.service.test.consul`

#### service_with_tag

Also see [service](#service) documentation above.

`[[ service_with_tag "test" "tag" ]]` will output `tag.test.service.consul`

Given `--variable consul_domain=test.consul`

`[[ service_with_tag "test" "tag" ]]` will output `tag.test.service.test.consul`

#### grant_credentials

Helper to output a policy path for credentials access.

This is useful for `database` and `rabbitmq` access policies.

`[[ grant_credentials "db-test" "full" ]]` will output

```hcl
path "db-test/creds/full" {
  capabilities = ["read"]
}
```

#### grant_credentials_policy

Helper to output a full policy (with path) for credentials access.

This is useful for `database` and `rabbitmq` access policies.

`[[ grant_credentials_policy "db-test" "full" ]]` will output

```hcl
policy "db-test-full" {
  path "db-test/creds/full" {
    capabilities = ["read"]
  }
}
```

#### github_assign_team_policy

Helper to output a github team to vault policy mapping.

`[[ github_assign_team_policy "infra" "infra-policy" ]]` will output

```hcl
secret "/auth/github/map/teams/infra" {
  value = "infra-policy"
}
```

#### ldap_assign_group_policy

Helper to output a ldap group to vault policy mapping.

`[[ ldap_assign_group_policy "infra" "infra-policy" ]]` will output

```hcl
secret "/auth/ldap/groups/infra" {
  value = "infra-policy"
}
```

### Example

#### HCL Variable file

```hcl
environment_name = "staging"
environment_tld  = "stag"

db_default_ttl = "9h"
db_max_ttl     = "72h"

mysql_databases = [
  "db-1",
  "db-2",
  "db-3"
]

mysql_irregular_database_names = {
  db-1 = "some_other_db"
}

mysql_role_full = <<-SQL
  CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';
  GRANT ALL ON __DB__.* TO '{{name}}'@'%';
  SQL

mysql_role_read_only = <<-SQL
  CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';
  GRANT SELECT ON __DB__.* TO '{{name}}'@'%';
  SQL

mysql_role_read_write = <<-SQL
  CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';
  GRANT SELECT, INSERT, UPDATE, DELETE, EXECUTE, SHOW VIEW, CREATE TEMPORARY TABLES, LOCK TABLES ON __DB__.* TO '{{name}}'@'%';
  GRANT PROCESS ON *.* TO '{{name}}'@'%';
  SQL
```

#### HCL template file

```hcl
environment "*" {
[[ range $k, $v := .mysql_databases ]]
  [[- $name := printf "db-%s" $v -]]
  [[- $environment_name := (lookup "environment_name") -]]

  mount "[[ $name ]]" {
    role "full" {
      db_name             = "default"
      default_ttl         = "[[ lookup "db_default_ttl" ]]"
      max_ttl             = "[[ lookup "db_max_ttl" ]]"
      creation_statements = <<-SQL
        [[ lookup "mysql_role_full" | replace_all "__DB__" (lookup_map_default "mysql_irregular_database_names" $v $v ) ]]
      SQL
    }

    role "read-write" {
      db_name             = "default"
      default_ttl         = "[[ lookup "db_default_ttl" ]]"
      max_ttl             = "[[ lookup "db_max_ttl" ]]"
      creation_statements = <<-SQL
        [[ lookup "mysql_role_read_write" | replace_all "__DB__" (lookup_map_default "mysql_irregular_database_names" $v $v ) ]]
      SQL
    }

    role "read-only" {
      db_name             = "default"
      default_ttl         = "[[ lookup "db_default_ttl" ]]"
      max_ttl             = "[[ lookup "db_max_ttl" ]]"
      creation_statements = <<-SQL
        [[ lookup "mysql_role_read_only" | replace_all "__DB__" (lookup_map_default "mysql_irregular_database_names" $v $v ) ]]
      SQL
    }
  }

  # grant "full" access policies for "[[ $name ]]"
  [[ grant_credentials_policy $name "full" ]]
  [[ github_assign_team_policy (printf "rds-%s-%s-full" $environment_name $name) (printf "%s-full" $name) ]]
  [[ ldap_assign_group_policy (printf "rds-%s-%s-full" $environment_name $name) (printf "%s-full" $name) ]]

  # grant "read-write" access policies for "[[ $name ]]"
  [[ grant_credentials_policy $name "read-write" ]]
  [[ github_assign_team_policy (printf "rds-%s-%s-read-write" $environment_name $name) (printf "%s-read-write" $name) ]]
  [[ ldap_assign_group_policy (printf "rds-%s-%s-read-write" $environment_name $name) (printf "%s-read-write" $name) ]]

  # grant "read-only" access policies for "[[ $name ]]"
  [[ grant_credentials_policy $name "read-only" ]]
  [[ github_assign_team_policy (printf "rds-%s-%s-read-only" $environment_name $name) (printf "%s-read-only" $name) ]]
  [[ ldap_assign_group_policy (printf "rds-%s-%s-read-only" $environment_name $name) (printf "%s-read-only" $name) ]]
[[ end ]]
}
```

## Workflow

The following is a sample workflow that may be used for organizations with Consul and Vault clusters in different environments. If your setup deviates from said description, feel free to modify your workflow.

### Directory Structure

The directory structure is laid out like described below:

- `/${env}/apps/${app}.hcl` (encrypted) Vault secrets or (cleartext) Consul KeyValue for an application in a specific environment.
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

### Consul app KV

```hcl
environment "production" {
  kv "name" "production" {}

  # application name must match the file name
  application "api-admin" {

    # cleartext shorthand configuration for the application, will be written to /api-admin/threads
    kv "threads" "10" {}

    # cleartext configuration for the application, will be written to /api-admin/config
    kv "config" {
      value = <<EOF
Some
file
value!
EOF
    }
  }
}
```

### Vault auth


```hcl
environment "production" {

  # The auth "name" can be anything, will be the "path" in auth configuration, e.g. the mount below will
  # make the "aws-ec2" secret backend available at "aws-ec2/" in the Vault API and CLI.
  #
  # The auth stanza maps to the Mount API ( https://www.vaultproject.io/api/system/auth.html#mount-auth-backend )
  # API endpoint: /sys/auth/:path (/sys/auth/aws-ec2)
  auth "aws-ec2" { # :path

    # Type must match the Vault auth types
    # based on this type, all config and role config below will map to settings found at https://www.vaultproject.io/docs/auth/aws.html
    type = "aws-ec2"

    # Client Configuration for the autb backend
    #
    # maps to the secret backend specific configuration
    # in this example it will be https://www.vaultproject.io/docs/auth/aws.html#auth-aws-config-client
    # key/value here is arbitrary and backend dependent, matches the Vaults docs 1:1 in keys and values
    #
    # API endpoint: /auth/:path/config/:config_name (/auth/aws-ec2/config/client)
    config "client" { # :config_name
      access_key = "XXXX"
      secret_key = "YYYY"
      max_ttl    = "1d"
    }

    # Auth backend type specific roles
    #
    # in this case it maps to https://www.vaultproject.io/docs/auth/aws.html#auth-aws-role-role-
    # key/value here is arbritary and backend dependent, matches the Vault docs 1:1 in keys and values
    #
    # API endpoint: /auth/:path/role/:role: (/auth/aws-ec2/role/api-admin-prod)
    role "api-admin-prod" { # :role
      policies                       = "global,sample-policy"
      max_ttl                        = "1h"
      allow_instance_migration       = false
      bound_vpc_id                   = "vpc-XXXXXX"
      bound_iam_instance_profile_arn = "arn:aws:iam::XXXXXXX:instance-profile/XXXX"
    }
  }

  auth "sg-github" {
    type = "github"

    # https://www.vaultproject.io/docs/auth/github.html#generate-a-github-personal-access-token
    #
    # API endpoint: auth/:name/config (auth/sg-github/config)
    config "" {
      organization = "seatgeek"
    }
  }
}
```

### Consul services

The following can be stored in a cleartext file at `production/consul_services/cache.hcl`

```hcl
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

```hcl
environment "production" {

  # the name / path the mount will be mounted at, must be unique for the environment
  mount "db-api" {

    # the mount backend type to use (see Vault docs)
    type = "database"

    # optional string
    max_lease_ttl     = "1h"

    # optional string
    default_lease_ttl = "1h"

    # optional boolean
    force_no_cache    = true

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
