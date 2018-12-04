# If configured, the value will be used for constructing Consul service
# hostnames during templating. The default value is just "consul"
#
consul_domain = "consul.stag"

# The long name of the environment
environment_name = "staging"

# The default Time To Live for DB leases
db_default_ttl = "9h"

# The maximum Time To Live for DB leases
db_max_ttl = "72h"

# List of databases in our environment
mysql_databases = [
  "alpha",
  "beta",
  "bravo",
]

# usually our instance name matches the DB name but in some cases, this is not true, for those
# cases you can override the default name here
mysql_irregular_db_names = {
  beta = "some_other_db"
}
