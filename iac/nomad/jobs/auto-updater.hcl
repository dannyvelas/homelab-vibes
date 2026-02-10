job "auto-updater" {
  type        = "batch"
  datacenters = ["dc1"]

  periodic {
    cron             = "0 3 * * *"
    prohibit_overlap = true
    time_zone        = "UTC"
  }

  group "updater" {
    count = 1

    task "check-updates" {
      driver = "docker"

      config {
        image   = "docker:24-cli"
        command = "/bin/sh"
        args    = ["-c", "/local/update-check.sh"]

        volumes = [
          "/var/run/docker.sock:/var/run/docker.sock:ro",
        ]
      }

      template {
        destination = "local/update-check.sh"
        perms       = "755"
        data        = <<-EOF
          #!/bin/sh
          set -e

          echo "Starting image update check at $(date)"

          NOMAD_ADDR="${NOMAD_ADDR:-http://localhost:4646}"
          UPDATED=false

          for APP in plex sonarr radarr bazarr; do
            echo "Checking $APP for updates..."

            # Get current image from Nomad job
            CURRENT_IMAGE=$(nomad job inspect "$APP" 2>/dev/null | jq -r '.Job.TaskGroups[0].Tasks[0].Config.image' 2>/dev/null || echo "")

            if [ -z "$CURRENT_IMAGE" ]; then
              echo "  Could not determine current image for $APP — skipping"
              continue
            fi

            echo "  Current image: $CURRENT_IMAGE"

            # Pull latest version
            docker pull "$CURRENT_IMAGE" > /dev/null 2>&1

            # Get the running container's image ID
            CONTAINER_ID=$(docker ps --filter "name=$APP" -q 2>/dev/null | head -1)
            if [ -z "$CONTAINER_ID" ]; then
              echo "  No running container for $APP — skipping digest check"
              continue
            fi

            RUNNING_DIGEST=$(docker inspect --format='{{.Image}}' "$CONTAINER_ID" 2>/dev/null || echo "")
            LATEST_DIGEST=$(docker inspect --format='{{.Id}}' "$CURRENT_IMAGE" 2>/dev/null || echo "")

            if [ -z "$RUNNING_DIGEST" ] || [ -z "$LATEST_DIGEST" ]; then
              echo "  Could not determine digests for $APP — skipping"
              continue
            fi

            if [ "$RUNNING_DIGEST" != "$LATEST_DIGEST" ]; then
              echo "  New version available for $APP — triggering deployment..."
              nomad job run "/generated/nomad/${APP}.hcl" 2>/dev/null && {
                echo "  $APP update triggered successfully"
                UPDATED=true
              } || echo "  WARNING: Failed to trigger update for $APP"
            else
              echo "  $APP is up to date"
            fi

            echo ""
          done

          if [ "$UPDATED" = true ]; then
            echo "Update check completed — updates applied at $(date)"
          else
            echo "Update check completed — no updates needed at $(date)"
          fi
        EOF
      }

      env {
        NOMAD_ADDR = "http://${attr.unique.network.ip-address}:4646"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }

    restart {
      attempts = 2
      delay    = "30s"
      interval = "5m"
      mode     = "fail"
    }
  }
}
