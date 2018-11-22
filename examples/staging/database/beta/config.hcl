environment "staging" {
  mount "db-beta" {
    type = "database"

    config "default" {
      plugin_name    = "mysql-rds-database-plugin"
      connection_url = "beta:beta@tcp(master.db-beta.service.stag.consul:3306)/"
      allowed_roles  = "*"
    }
  }
}
