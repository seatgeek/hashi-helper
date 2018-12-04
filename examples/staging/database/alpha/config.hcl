environment "staging" {
  mount "db-alpha" {
    type = "database"

    config "default" {
      plugin_name    = "mysql-rds-database-plugin"
      connection_url = "alpha:alpha@tcp(master.db-alpha.service.stag.consul:3306)/"
      allowed_roles  = "*"
    }
  }
}
