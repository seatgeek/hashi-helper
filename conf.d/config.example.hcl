environment "production" {
  application "peakpass" {
    secret "DATABASE_URL" {
      value = "production-peakpass-DATABASE_URL"
    }

    secret "REDIS_URL" {
      value = "production-peakpass-REDIS_URL"
    }

    secret "S3_BUCKET" {
      value = "production-peakpass-S3_BUCKET"
    }
  }
}

environment "staging" {
  application "peakpass" {
    secret "DATABASE_URL" {
      value = "staging-peakpass-DATABASE_URL"
    }

    secret "REDIS_URL" {
      value = "staging-peakpass-REDIS_URL"
    }

    secret "S3_BUCKET" {
      value = "staging-peakpass-S3_BUCKET"
    }
  }
}

environment "production" {
  application "peakpass" {
    secret "DATABASE_URL" {
      value = "derpie derp"
    }
  }
}
