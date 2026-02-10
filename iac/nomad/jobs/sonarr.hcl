job "sonarr" {
  type        = "service"
  datacenters = ["dc1"]

  group "sonarr" {
    count = 1

    network {
      port "sonarr" {
        static = 8989
      }
    }

    service {
      name = "sonarr"
      port = "sonarr"

      check {
        name     = "sonarr-health"
        type     = "http"
        path     = "/health"
        port     = "sonarr"
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

    task "sonarr" {
      driver = "docker"

      config {
        image = "linuxserver/sonarr:latest"
        ports = ["sonarr"]

        readonly_rootfs = true

        tmpfs = [
          "/tmp",
        ]

        volumes = [
          "/mnt/config/sonarr:/config",
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
