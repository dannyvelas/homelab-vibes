better differences:
- has fail2ban
- explicitly set ssh pub key to yet
- made directory for playbooks
- used raw iptable logic instead of just ufw
- seems to have more firewall rules
- split responsibility of creating VM and creating plex. in my project, creating a VM was the responsibility of the plex playbook. here, creating VMs is always the responsibility of the "configure host" playbook. the plex playbook just has to find the vm and install itself in it, like in k8s
- using wireguard instead of tailscale
- using a reverse proxy like Traefik
- adding ddns in case my home-ip changes
- having a "status" sub-command to see what apps are running for all hosts
- having a "security-audit" sub-command and playbook 
- having a "terraform teardown", although i think i was going to have this anyway
- running terraform init always when terraform runs. it is an idempotent command that will work for the very first time and every time after even if it's unecessary
- generates ansible inventory idempotently on every run instead of assuming one exists. also uses yml for inventory which is easier to parse and update programmatically than .ini files.

maybe better, maybe worse differences:
- creating `group_vars` directory with variables for ansible instead of passing in variables via command line via json.
  - Turns out my approach was better, but AI was right about one thing: instead of using a temp file, i could keep the generated config file inside of a git-ignored ".generated" directory for easier debugging.
- creating `*.tfvars` files for terraform instead of passing in variables via command line
  - same here, my approach was better, but AI was right about one thing: instead of using a temp file, i can keep the generated config file inside of a git-ignored ".generated" directory for easier debugging.
- my project has a more granular CLI, where to set up something that under-the-hood necessitates both terraform and ansible logic, you would have to run one or more commands for terraform `iac terraform ...args` and one or more commands for ansible `iac ansible ...args`. if you wanted to have one command that sets up everything under-the-hood you would have to use a task runner like Taskfile or makefile as a layer of abstraction. this AI project wrote the code so that you don't have the same granular control which would allow you to only run only the terraform part of setting something up. it only allows you to set up the full thing `iac provision ....`.
  - this ai approach could be worse since you can't be as granular as my approach
  - but it could be better. an argument could be made that being granular doesn't matter, and that there are limitations that a task runner will face, that won't happen if you have everything abstracted behind one go program. e.g. one limitation is that you can't pass data (or it is harder / more awkward) to pass context data from one execution of terraform to an execution of ansible in my approach. this would have to be passed via CLI in my approach. but in the programming approach it's trivial to pass context data from the execution of terraform to an execution of ansible in this approach because both will be executed within the same go function.
  - too early to tell, i will continue with my approach because i like the granularity and hopefully i don't run into a case where i need to pass context from one cli command to the other via taskfile. maybe if i find myself needing to pass context from one command to another it will be an indication to me that it is too granular and i can just merge those two commands. maybe this philosophy is enough to get the best of both worlds: granular CLI and no context passing
- the AI uses libvirt/KVM terraform provider instead of incus
- the AI sets "allow agent forwarding" to no for ssh
  - this just makes it so you cant "ssh hop" from one machine to another
- the AI sets "max auth tries" to 3 for SSH
  - this helps prevent brute force attacks
- the AI sets "x11 forwarding" to no for ssh
  - this is only relevant if you are using the X window system on a server and want to prevent ssh forwarding from the GUI. but both of my servers don't have a window system installed on them. so this does nothing.
- the AI created a dedicated task to enable ufw, where in mine it seems implied
  - mine is fine, it actually does explicitly enable UFW in one task, but it had a misleading title, so i fixed the title
- the AI uses handlers instead of restarting sshd as a task
  - i told it my approach and it seems to like mine more so i'll keep it

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
- [ ] doesn't email when an update happens
- [ ] doesn't create SSH user for VM. is this necessary? i thought yes to be able to run an ansible playbook on it
- [ ] the CLI might be broken. it looks like it asks you to set up one VM at a time. but when you run it for only one VM i think it sets up everything anyway.
- [ ] it does have a way for you to be able to see the configs that are being used. this is in a way better than my current cli behavior. my cli only shows whether a config is loaded or not. it doesn't show you the full value. however, my cli might be better in the sense that it only shows the exact minimum configs needed for a command. if configs are say, $10^9$ keys, and a command you're interested in running only uses 5, you'd much rather see output that tells you whether those 5 configs are set correctly, rather than output of $10^9$ keys. my CLI can be improved, but i think it's easier to improve than the AI approach is.

questions:
- [x] do we actually need both computers?
    - it looks like we’re actually deploying plex and sonar to one computer and radarr and bazarr to the other computer. let’s see how much capacity we have left 
    - according to claude, there is roughly 2.8-3.3 GB free per VM and ~2.7 GB free per host. That's enough for Atlantis or OVN individually, and probably both.
