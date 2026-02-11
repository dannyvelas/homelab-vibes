fixes:
- [ ] fix `findRepoRoot` function, doesn't actually find repo root. doesn't work if you run it inside of iac/cli directory. returns `"/Users/dannyvelasquez/RemoteGit/MyGithub/homelab-vibe/iac/cli"` instead of `"/Users/dannyvelasquez/RemoteGit/MyGithub/homelab-vibe/"`.

future:
- [ ] add some TACO software (like terraform enterprise, scalr, spacelift, env0) so that terraform repo is always the absolute source of truth
  - looks liks terraform enterprise is too heavy
  - scalr, spacelift, env0 are SAAS
  - so can't do this yet
  - so, only options are to wait for more capacity or use atlantis
- [x] add OVN. on adding it, should we remove the UFW stuff?
- [x] do we actually need both computers?
    - it looks like we’re actually deploying plex and sonar to one computer and radarr and bazarr to the other computer. let’s see how much capacity we have left 
    - according to claude, there is roughly 2.8-3.3 GB free per VM and ~2.7 GB free per host. That's enough for Atlantis or OVN individually, and probably both.
- [ ] let’s stop making plex/radarr/sonarr/bazarr special. instead, let’s our system generic enough so that we can just deploy these as if they were any arbitrary dockerized service
- [ ] we need to expose plex
