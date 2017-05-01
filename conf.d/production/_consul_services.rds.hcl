environment "production" {
  service "db-tzanalytic" {
    id      = "db-tzanalytic-master"
    node    = "rds"
    address = "production-tzanalytic.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master"]
  }

  service "db-tzanalytic" {
    id      = "db-tzanalytic-replica"
    node    = "rds"
    address = "production-tzanalytic-replica.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["replica"]
  }

  service "db-sellerdirect" {
    id      = "db-sellerdirect-master"
    node    = "rds"
    address = "production-sellerdirect.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master"]
  }

  service "db-sellerdirect" {
    id      = "db-sellerdirect-replica"
    node    = "rds"
    address = "production-sellerdirect-read-replica.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["replica"]
  }

  service "db-uberseat" {
    id      = "db-uberseat-master"
    node    = "rds"
    address = "production-uberseat.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master"]
  }

  service "db-uberseat" {
    id      = "db-uberseat-replica"
    node    = "rds"
    address = "production-uberseat-read-replica.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["replica"]
  }

  service "db-peakpass" {
    id      = "db-peakpass-master"
    node    = "rds"
    address = "production-peakpass.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 5432
    tags    = ["master"]
  }

  service "db-peakpass" {
    id      = "db-peakpass-replica"
    node    = "rds"
    address = "production-peakpass-replica.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 5432
    tags    = ["replica"]
  }

  service "db-paulbunyan" {
    id      = "db-peakpass-master"
    node    = "rds"
    address = "production-paulbunyan.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master"]
  }

  service "db-paulbunyan" {
    id      = "db-paulbunyan-replica"
    node    = "rds"
    address = "production-paulbunyan-read-replica.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["replica"]
  }

  service "db-alerts" {
    id      = "db-alerts-master"
    node    = "rds"
    address = "production-alerts.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master"]
  }

  service "db-alerts" {
    id      = "db-alerts-replica"
    node    = "rds"
    address = "production-alerts-read-replica.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["replica"]
  }

  service "db-apollo" {
    id      = "db-apollo"
    node    = "rds"
    address = "production-apollo.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }

  service "db-blart" {
    id      = "db-blart"
    node    = "rds"
    address = "production-blart.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }

  service "db-blog" {
    id      = "db-blog"
    node    = "rds"
    address = "production-blog.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }

  service "db-buster" {
    id      = "db-buster"
    node    = "rds"
    address = "production-buster-pg95.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 5432
    tags    = ["master", "replica"]
  }

  service "db-buzzfeed" {
    id      = "db-buzzfeed"
    node    = "rds"
    address = "production-buzzfeed.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 5432
    tags    = ["master", "replica"]
  }

  service "db-cronq" {
    id      = "db-cronq"
    node    = "rds"
    address = "production-cronq.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }

  service "db-cronq" {
    id      = "db-cronq"
    node    = "rds"
    address = "production-cronq.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }

  service "db-dharma" {
    id      = "db-dharma"
    node    = "rds"
    address = "production-dharma.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }

  service "db-inferno" {
    id      = "db-inferno"
    node    = "rds"
    address = "production-inferno.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 5432
    tags    = ["master", "replica"]
  }

  service "db-janestreet" {
    id      = "db-janestreet"
    node    = "rds"
    address = "production-janestreet.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }

  service "db-manbearpig" {
    id      = "db-manbearpig"
    node    = "rds"
    address = "production-manbearpig.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }

  service "db-sevenpack" {
    id      = "db-sevenpack"
    node    = "rds"
    address = "production-sevenpack.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }

  service "db-sirius" {
    id      = "db-sirius"
    node    = "rds"
    address = "production-sirius.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 5432
    tags    = ["master", "replica"]
  }

  service "db-tracker" {
    id      = "db-tracker"
    node    = "rds"
    address = "production-tracker.ccgmbht45ta2.us-east-1.rds.amazonaws.com"
    port    = 3306
    tags    = ["master", "replica"]
  }
}
