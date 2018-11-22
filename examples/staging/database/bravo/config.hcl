environment "staging" {
  mount "db-bravo" {
    type = "database"

    config "default" {
      plugin_name    = "mysql-rds-database-plugin"
      connection_url = "bravo:bravo@tcp(master.db-bravo.service.stag.consul:3306)/"
      allowed_roles  = "*"
    }
  }
}
