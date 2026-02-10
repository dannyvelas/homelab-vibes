future:
- add some TACO software (like terraform enterprise, scalr, spacelift, env0) so that terraform repo is always the absolute source of truth
- add OVN. do we have capacity for it? if so, should we remove the UFW stuff?
- do we actually need both computers?
    - it looks like we’re actually deploying plex and sonar to one computer and radarr and bazarr to the other computer. let’s see how much capacity we have left 
- let’s stop making plex/radarr/sonarr/bazarr special. instead, let’s our system generic enough so that we can just install these
- we need to expose plex
