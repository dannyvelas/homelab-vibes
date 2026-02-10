job "proxy" {
  type        = "service"
  datacenters = ["dc1"]

  group "traefik" {
    count = 1

    network {
      port "http" {
        static = 80
      }
      port "https" {
        static = 443
      }
    }

    service {
      name = "traefik"
      port = "https"

      check {
        name     = "traefik-ping"
        type     = "http"
        path     = "/ping"
        port     = "http"
        interval = "10s"
        timeout  = "2s"
      }
    }

    task "traefik" {
      driver = "docker"

      config {
        image = "traefik:v3.0"
        ports = ["http", "https"]

        volumes = [
          "local/traefik.yml:/etc/traefik/traefik.yml:ro",
          "local/dynamic.yml:/etc/traefik/dynamic.yml:ro",
        ]
      }

      template {
        destination = "local/traefik.yml"
        data        = <<-EOF
          api:
            dashboard: false
            ping: {}

          entryPoints:
            web:
              address: ":80"
              http:
                redirections:
                  entryPoint:
                    to: websecure
                    scheme: https
                    permanent: true

            websecure:
              address: ":443"
              http:
                tls: {}
                middlewares:
                  - securityHeaders@file
                  - rateLimit@file

          providers:
            file:
              filename: /etc/traefik/dynamic.yml
              watch: true

          log:
            level: INFO
        EOF
      }

      template {
        destination = "local/dynamic.yml"
        data        = <<-EOF
          tls:
            options:
              default:
                minVersion: VersionTLS12
                cipherSuites:
                  - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
                  - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
                  - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305
                  - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305

          http:
            middlewares:
              securityHeaders:
                headers:
                  stsSeconds: 31536000
                  stsIncludeSubdomains: true
                  stsPreload: true
                  forceSTSHeader: true
                  frameDeny: true
                  customFrameOptionsValue: "DENY"
                  contentTypeNosniff: true
                  browserXssFilter: true
                  referrerPolicy: "strict-origin-when-cross-origin"
                  customResponseHeaders:
                    X-Frame-Options: "DENY"
                    Content-Security-Policy: >-
                      default-src 'self';
                      script-src 'self' 'unsafe-inline';
                      style-src 'self' 'unsafe-inline';
                      img-src 'self' data: https:;
                      connect-src 'self';
                      frame-ancestors 'none'

              rateLimit:
                rateLimit:
                  average: 100
                  burst: 50
                  period: 1s
                  sourceCriterion:
                    ipStrategy:
                      depth: 1

            routers:
              plex:
                rule: "PathPrefix(`/plex`)"
                entryPoints:
                  - websecure
                middlewares:
                  - securityHeaders
                  - rateLimit
                service: plex
                tls: {}

              sonarr:
                rule: "PathPrefix(`/sonarr`)"
                entryPoints:
                  - websecure
                middlewares:
                  - securityHeaders
                  - rateLimit
                service: sonarr
                tls: {}

              radarr:
                rule: "PathPrefix(`/radarr`)"
                entryPoints:
                  - websecure
                middlewares:
                  - securityHeaders
                  - rateLimit
                service: radarr
                tls: {}

              bazarr:
                rule: "PathPrefix(`/bazarr`)"
                entryPoints:
                  - websecure
                middlewares:
                  - securityHeaders
                  - rateLimit
                service: bazarr
                tls: {}

            services:
              plex:
                loadBalancer:
                  servers:
                    - url: "http://127.0.0.1:32400"

              sonarr:
                loadBalancer:
                  servers:
                    - url: "http://127.0.0.1:8989"

              radarr:
                loadBalancer:
                  servers:
                    - url: "http://127.0.0.1:7878"

              bazarr:
                loadBalancer:
                  servers:
                    - url: "http://127.0.0.1:6767"
        EOF
      }

      resources {
        cpu    = 200
        memory = 256
      }
    }
  }
}
