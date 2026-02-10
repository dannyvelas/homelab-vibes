job "radarr" {
  type        = "service"
  datacenters = ["dc1"]

  group "radarr" {
    count = 1

    network {
      port "radarr" {
        static = 7878
      }
    }

    service {
      name = "radarr"
      port = "radarr"

      check {
        name     = "radarr-health"
        type     = "http"
        path     = "/health"
        port     = "radarr"
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

    task "radarr" {
      driver = "docker"

      config {
        image = "linuxserver/radarr:latest"
        ports = ["radarr"]

        readonly_rootfs = true

        tmpfs = [
          "/tmp",
        ]

        volumes = [
          "/mnt/config/radarr:/config",
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
