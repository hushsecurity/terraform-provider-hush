# Create an access policy for a PostgreSQL credential
resource "hush_access_policy" "postgres_example" {
  name                 = "prod-db-policy"
  description          = "Access policy for production database"
  enabled              = true
  access_credential_id = hush_postgres_access_credential.example.id
  access_privilege_ids = [hush_postgres_access_privilege.example.id]
  deployment_ids       = [hush_deployment.example.id]

  attestation_criteria {
    type  = "k8s:ns"
    value = "production"
  }

  attestation_criteria {
    type  = "k8s:sa"
    value = "app-service-account"
  }

  # Template: postgresql://${username}:${password}@${host}:${port}/${db}
  env_delivery_config {
    item {
      name = "DATABASE_URL"
      type = "template"
      key  = "postgresql://$${username}:$${password}@$${host}:$${port}/$${db}"
    }
  }
}

# Create an access policy for a MongoDB credential
resource "hush_access_policy" "mongodb_example" {
  name                 = "prod-mongodb-policy"
  description          = "Access policy for production MongoDB"
  enabled              = true
  access_credential_id = hush_mongodb_access_credential.example.id
  deployment_ids       = [hush_deployment.example.id]

  attestation_criteria {
    type  = "k8s:ns"
    value = "production"
  }

  # Template: mongodb://${username}:${password}@${host}:${port}/${db_name}?authSource=${auth_source}
  env_delivery_config {
    item {
      name = "MONGODB_URI"
      type = "template"
      key  = "mongodb://$${username}:$${password}@$${host}:$${port}/$${db_name}?authSource=$${auth_source}"
    }
  }
}

# Create an access policy for a MySQL credential
resource "hush_access_policy" "mysql_example" {
  name                 = "prod-mysql-policy"
  description          = "Access policy for production MySQL"
  enabled              = true
  access_credential_id = hush_mysql_access_credential.example.id
  access_privilege_ids = [hush_mysql_access_privilege.example.id]
  deployment_ids       = [hush_deployment.example.id]

  attestation_criteria {
    type  = "k8s:ns"
    value = "production"
  }

  # Template: mysql://${username}:${password}@${host}:${port}/${db_name}
  env_delivery_config {
    item {
      name = "MYSQL_URL"
      type = "template"
      key  = "mysql://$${username}:$${password}@$${host}:$${port}/$${db_name}"
    }
  }
}

# Create an access policy for a MariaDB credential
resource "hush_access_policy" "mariadb_example" {
  name                 = "prod-mariadb-policy"
  description          = "Access policy for production MariaDB"
  enabled              = true
  access_credential_id = hush_mariadb_access_credential.example.id
  deployment_ids       = [hush_deployment.example.id]

  attestation_criteria {
    type  = "k8s:ns"
    value = "production"
  }

  # Template: mariadb://${username}:${password}@${host}:${port}/${db_name}
  env_delivery_config {
    item {
      name = "MARIADB_URL"
      type = "template"
      key  = "mariadb://$${username}:$${password}@$${host}:$${port}/$${db_name}"
    }
  }
}

# Create an access policy for an OpenAI credential
resource "hush_access_policy" "openai_example" {
  name                 = "prod-openai-policy"
  description          = "Access policy for OpenAI API"
  enabled              = true
  access_credential_id = hush_openai_access_credential.example.id
  access_privilege_ids = [hush_openai_access_privilege.example.id]
  deployment_ids       = [hush_deployment.example.id]

  attestation_criteria {
    type  = "k8s:ns"
    value = "production"
  }

  env_delivery_config {
    item {
      name = "OPENAI_API_KEY"
      key  = "api_key"
    }
  }
}

# Create an access policy for a Gemini credential
resource "hush_access_policy" "gemini_example" {
  name                 = "prod-gemini-policy"
  description          = "Access policy for Gemini API"
  enabled              = true
  access_credential_id = hush_gemini_access_credential.example.id
  deployment_ids       = [hush_deployment.example.id]

  attestation_criteria {
    type  = "k8s:ns"
    value = "production"
  }

  env_delivery_config {
    item {
      name = "GEMINI_SERVICE_ACCOUNT_KEY"
      key  = "service_account_key"
    }
  }
}
