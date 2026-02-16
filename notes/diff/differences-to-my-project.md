better differences:
- has fail2ban
- explicitly set ssh pub key to yet
- made directory for playbooks

maybe better, maybe worse differences:
- creating `group_vars` directory with variables for ansible instead of passing in variables via command line via json
- creating `*.tfvars` files for terraform instead of passing in variables via command line
- my project has a more granular CLI, where to set up something that under-the-hood necessitates both terraform and ansible logic, you would have to run one or more commands for terraform `iac terraform ...args` and one or more commands for ansible `iac ansible ...args`. if you wanted to have one command that sets up everything under-the-hood you would have to use a task runner like Taskfile or makefile as a layer of abstraction. this AI project wrote the code so that you don't have the same granular control which would allow you to only run only the terraform part of setting something up. it only allows you to set up the full thing `iac provision ....`.
  - this ai approach could be worse since you can't be as granular as my approach
  - but it could be better. an argument could be made that being granular doesn't matter, and that there are limitations that a task runner will face, that won't happen if you have everything abstracted behind one go program. e.g. one limitation is that you can't pass data (or it is harder / more awkward) to pass context data from one execution of terraform to an execution of ansible in my approach. this would have to be passed via CLI in my approach. but in the programming approach it's trivial to pass context data from the execution of terraform to an execution of ansible in this approach because both will be executed within the same go function.
- the AI uses libvirt/KVM terraform provider instead of incus
- the AI sets "allow agent forwarding" to no for ssh
- the AI sets "max auth tries" to 3 for SSH
- the AI sets "x11 forwarding" to no for ssh
- the AI created a dedicated task to enable ufw, where in mine it seems implied

worse differences:
- [ ] business requirement: being able to build iac easily. right now the instructions require you to cd into iac/cli, build, and then cd back out. this is annoying. 
  - add a taskfile with a task that has the command in the README. that way all you have to do is execute one task command and it will cd, build the binary, and cd back for you.
    - PROBLEM: actually the command in the README is broken. it creates the binary in the `iac` folder instead of in the root of the repo. so we could do this but the command would need to be fixed
    - PROBLEM: `go run` and `dlv debug` won't work at all; they will fail at the root of the project because the root of the project is not a go module. they will fail at the root of the go project because the code isn't designed to work unless you run it from the root of the project. for delve, you would have to compile first in the root of the go project and then use `dlv exec` in the root of the go repo, which is less-than-ideal.
  - fix code so that it can be run from either directory
    - PROBLEM: this will make `go run` and `dlv debug` work at the root of the go repo. but, they will still continue not working if this is run from the root of the project. this is because the root of the project is not a go module.
  - make go code just be at the top-level
- [ ] homelab.yml.example exists in root of project and also one exists in root of go repo. this is unnecessary duplication
- [ ] i don't like how some flags are optional and others are required list for `provision` `--host` is required but `--config` is optional. not intuitive which flags need to be passed in
- [ ] the way that it's changing the ssh permissions is suboptimal. its doing regex search and replace on `/etc/ssh/sshd_config` instead of just creating a new file that will get priority
- [ ] the libvirt terraform file was completely wrong :(
- [ ] it tries to run Terraform (which expects libvirt to exist on the machine) before running ansible (which is the thing that installs libvirt)
- [ ] doesn't switch to a random port

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
