job "bazarr" {
  type        = "service"
  datacenters = ["dc1"]

  group "bazarr" {
    count = 1

    network {
      port "bazarr" {
        static = 6767
      }
    }

    service {
      name = "bazarr"
      port = "bazarr"

      check {
        name     = "bazarr-health"
        type     = "http"
        path     = "/health"
        port     = "bazarr"
        interval = "30s"
        timeout  = "5s"
      }
    }

    update {
      max_parallel      = 1
      min_healthy_time  = "30s"
      healthy_deadline  = "5m"
      progress_deadline = "10m"
      auto_revert       = true
      canary            = 0
    }

    task "bazarr" {
      driver = "docker"

      config {
        image = "linuxserver/bazarr:latest"
        ports = ["bazarr"]

        readonly_rootfs = true

        tmpfs = [
          "/tmp",
        ]

        volumes = [
          "/mnt/config/bazarr:/config",
          "/mnt/media:/media",
          "/mnt/downloads:/downloads",
        ]
      }

      env {
        PUID = "1000"
        PGID = "1000"
        TZ   = "America/Los_Angeles"
      }

      resources {
        cpu    = 500
        memory = 512
      }
    }
  }
}
