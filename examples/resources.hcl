environment "env-demo" {
}

resource "postgres" "database" {
  state = {
    engine = "16"
    storage_gb = 100
  }
}

resource "redis" "cache" {
  state = {
    memory_mb = 512
  }
  depends_on = ["database"]
}
