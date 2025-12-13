variable "dialect" {
  type    = string
  default = "postgres"
}

# Use the GORM provider directly (no loader program)
data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/models",
    "--dialect", var.dialect,
  ]
}

# Local development environment
env "local" {
  src = data.external_schema.gorm.url
  dev = getenv("NEON_DEV_DATABASE_URL")

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

# Staging environment
env "staging" {
  src = data.external_schema.gorm.url
  url = getenv("NEON_STAGING_DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}

# Production environment
env "production" {
  src = data.external_schema.gorm.url
  url = getenv("NEON_PROD_DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}