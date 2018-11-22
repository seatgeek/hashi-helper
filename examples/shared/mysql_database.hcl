# Example configuration template file
#
# Template files use https://golang.org/pkg/text/template/ for rendering, with the minor change
# of the use of "[[" and "]]" as delimiters instead of "{{" and "}}" which will overlap with HCLs
# own { } brackets
#

environment "*" {
# iterate all databases provided in staging/config.var.hcl
# each item in the list will be iterated one by one, and exposed as
# $name within the template
[[ range $name := .mysql_databases ]]

  # define an inline variable called "mount_name", constructed using the build-in
  # printf function
  [[ $mount_name := printf "db-%s" $name ]]

  # define an inline variabled called "database_name".
  # we are looking up the database name in "mysql_irregular_db_names" (from staging/config.var.hcl)
  # and prvoide "$name" as the default value if the key do not exist in the map
  [[ $database_name := lookupMapDefault "mysql_irregular_db_names" $name $name ]]

  # define an inline variable called "environment_name" which simply look up
  # the key "environment_name" in staging/config.var.hcl
  [[ $environment_name := (lookup "environment_name") ]]

  mount "[[ $mount_name ]]" {
    role "full" {
      db_name             = "default"
      default_ttl         = "[[ lookup "db_default_ttl" ]]"
      max_ttl             = "[[ lookup "db_max_ttl" ]]"
      creation_statements = <<-SQL
        CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';
        GRANT ALL ON [[ $database_name ]].* TO '{{name}}'@'%';
      SQL
    }

    role "read-write" {
      db_name             = "default"
      default_ttl         = "[[ lookup "db_default_ttl" ]]"
      max_ttl             = "[[ lookup "db_max_ttl" ]]"
      creation_statements = <<-SQL
        CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';
        GRANT SELECT, INSERT, UPDATE, DELETE, EXECUTE, SHOW VIEW, CREATE TEMPORARY TABLES, LOCK TABLES ON [[ $database_name ]].* TO '{{name}}'@'%';
        GRANT PROCESS ON *.* TO '{{name}}'@'%';
      SQL
    }

    role "read-only" {
      db_name             = "default"
      default_ttl         = "[[ lookup "db_default_ttl" ]]"
      max_ttl             = "[[ lookup "db_max_ttl" ]]"
      creation_statements = <<-SQL
        CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';
        GRANT SELECT ON [[ $database_name ]].* TO '{{name}}'@'%';
      SQL
    }
  }

  # grant "full" access policies for "[[ $mount_name ]]"
  [[ grantCredentialsPolicy $mount_name "full" ]]
  [[ githubAssignTeamPolicy (printf "rds-%s-%s-full" $environment_name $mount_name) (printf "%s-full" $mount_name) ]]
  [[ ldapAssignGroupPolicy (printf "rds-%s-%s-full" $environment_name $mount_name) (printf "%s-full" $mount_name) ]]

  # grant "read-write" access policies for "[[ $mount_name ]]"
  [[ grantCredentialsPolicy $mount_name "read-write" ]]
  [[ githubAssignTeamPolicy (printf "rds-%s-%s-read-write" $environment_name $mount_name) (printf "%s-read-write" $mount_name) ]]
  [[ ldapAssignGroupPolicy (printf "rds-%s-%s-read-write" $environment_name $mount_name) (printf "%s-read-write" $mount_name) ]]

  # grant "read-only" access policies for "[[ $mount_name ]]"
  [[ grantCredentialsPolicy $mount_name "read-only" ]]
  [[ githubAssignTeamPolicy (printf "rds-%s-%s-read-only" $environment_name $mount_name) (printf "%s-read-only" $mount_name) ]]
  [[ ldapAssignGroupPolicy (printf "rds-%s-%s-read-only" $environment_name $mount_name) (printf "%s-read-only" $mount_name) ]]
[[ end ]]
}
