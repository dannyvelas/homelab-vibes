job "plex" {
  type        = "service"
  datacenters = ["dc1"]

  group "plex" {
    count = 1

    network {
      port "plex" {
        static = 32400
      }
    }

    service {
      name = "plex"
      port = "plex"

      check {
        name     = "plex-health"
        type     = "http"
        path     = "/web/index.html"
        port     = "plex"
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

    task "plex" {
      driver = "docker"

      config {
        image = "linuxserver/plex:latest"
        ports = ["plex"]

        readonly_rootfs = true

        tmpfs = [
          "/tmp",
        ]

        volumes = [
          "/mnt/config/plex:/config",
          "/mnt/media:/media:ro",
          "/mnt/downloads:/downloads",
        ]
      }

      env {
        PUID     = "1000"
        PGID     = "1000"
        TZ       = "America/Los_Angeles"
        VERSION  = "docker"
      }

      resources {
        cpu    = 1000
        memory = 1024
      }
    }
  }
}
