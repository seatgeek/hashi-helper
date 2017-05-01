environment "production" {
  service "cache-santamaria" {
    id      = "cache-santamaria"
    node    = "cache"
    address = "prod-santamaria.ang13m.0001.use1.cache.amazonaws.com"
    port    = 6379
    tags    = ["master", "replica"]
  }

  service "cache-seatgeekapp" {
    id      = "cache-seatgeekapp"
    node    = "cache"
    address = "prod-seatgeekapp.ang13m.0001.use1.cache.amazonaws.com"
    port    = 6379
    tags    = ["master", "replica"]
  }

  service "cache-sellerdirect" {
    id      = "cache-sellerdirect"
    node    = "cache"
    address = "prod-sellerdirect.ang13m.0001.use1.cache.amazonaws.com"
    port    = 6379
    tags    = ["master", "replica"]
  }

  service "cache-shared" {
    id      = "cache-shared"
    node    = "cache"
    address = "prod-shared.ang13m.0001.use1.cache.amazonaws.com"
    port    = 6379
    tags    = ["master", "replica"]
  }

  service "cache-peakpass" {
    id      = "cache-peakpass-001"
    node    = "cache"
    address = "prod-peakpass-001.ang13m.0001.use1.cache.amazonaws.com"
    port    = 6379
    tags    = ["master"]
  }

  service "cache-peakpass" {
    id      = "cache-peakpass-002"
    node    = "cache"
    address = "prod-peakpass-002.ang13m.0001.use1.cache.amazonaws.com"
    port    = 6379
    tags    = ["replica"]
  }

  service "cache-sixpack" {
    id      = "cache-sixpack-001"
    node    = "cache"
    address = "prod-sixpack-001.ang13m.0001.use1.cache.amazonaws.com"
    port    = 6379
    tags    = ["replica"]
  }

  service "cache-sixpack" {
    id      = "cache-sixpack-002"
    node    = "cache"
    address = "prod-sixpack-002.ang13m.0001.use1.cache.amazonaws.com"
    port    = 6379
    tags    = ["master"]
  }
}
