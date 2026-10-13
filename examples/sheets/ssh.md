---
title: SSH
description: Log in to remote hosts, manage keys, tunnel ports and keep connections tidy with OpenSSH.
category: networking
tags: [ssh, openssh, keys, tunnels, remote]
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the OpenSSH 10.3p1 documentation; examples are displayed, never executed.
---
OpenSSH's `ssh` client opens encrypted sessions to remote machines. Most day-to-day convenience comes from three companions: `ssh-keygen` for keys, `ssh-agent`/`ssh-add` for unlocked keys, and `~/.ssh/config` for per-host defaults.

## Connect {.cols-2 #connect}
### Log in {#login}
```sh
ssh ada@build.example.org
ssh -p 2222 ada@build.example.org
ssh -i ~/.ssh/id_deploy deploy@web1.example.org
```

### Run a remote command {#remote-command}
```sh
ssh web1.example.org 'df -h /srv'
ssh -t web1.example.org htop
```

`-t` forces a terminal for interactive programs.

### Hop through a bastion {#jump}
```sh
ssh -J ada@bastion.example.org ada@db1.internal
```

`-J` connects through one or more comma-separated jump hosts.

### Debug a connection {.tip #verbose}
```sh
ssh -v build.example.org
ssh -G build.example.org | grep -i identityfile
```

`-v` (up to `-vvv`) logs the handshake. `-G` prints the effective configuration without connecting.

## Keys and agent {#keys}
### Create a key pair {#keygen}
```sh
ssh-keygen -t ed25519 -C "ada@laptop"
ssh-keygen -t ed25519 -f ~/.ssh/id_deploy
```

### Install your public key {#copy-id}
```sh
ssh-copy-id -i ~/.ssh/id_ed25519.pub ada@build.example.org
```

Appends the key to the remote `authorized_keys`, after logging in with an existing method.

### Inspect and maintain keys {.col-span-2 #key-maintenance}
| Command | Purpose |
| --- | --- |
| `ssh-keygen -l -f ~/.ssh/id_ed25519.pub` | Show a key's fingerprint |
| `ssh-keygen -y -f ~/.ssh/id_ed25519` | Print the public key for a private key |
| `ssh-keygen -p -f ~/.ssh/id_ed25519` | Change the passphrase |
| `ssh-keygen -R build.example.org` | Forget a host's stored key in `known_hosts` |

### Use the agent {#agent}
| Command | Purpose |
| --- | --- |
| `ssh-add ~/.ssh/id_ed25519` | Unlock a key for this session |
| `ssh-add -t 1h ~/.ssh/id_deploy` | Unlock for one hour |
| `ssh-add -l` | List loaded key fingerprints |
| `ssh-add -D` | Remove all keys from the agent |

### Agent forwarding {.warning #forward-agent}
```sh
ssh -A ada@bastion.example.org
```

Anyone with root on the remote host can use your forwarded agent while you are connected. Prefer `-J` for hopping.

## Tunnels {#tunnels}
### Local forward {.col-span-2 #local-forward}
```sh
ssh -N -L 5433:db1.internal:5432 ada@bastion.example.org
```

Local port 5433 now reaches `db1.internal:5432` as seen from the bastion. `-N` opens the tunnel without a remote shell.

### Remote forward {#remote-forward}
```sh
ssh -N -R 8080:localhost:3000 ada@demo.example.org
```

Port 8080 on the remote host connects back to your local port 3000.

### SOCKS proxy {#socks}
```sh
ssh -N -D 1080 ada@bastion.example.org
```

## Configuration {#config}
### Per-host defaults {.col-span-2 #host-blocks}
```text
Host build
    HostName build.example.org
    User ada
    Port 2222
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes

Host db1
    HostName db1.internal
    ProxyJump build
```

With this in `~/.ssh/config`, `ssh build` and `ssh db1` need no further options. The first matching value for each keyword wins.

### Reuse connections {#multiplex}
```text
Host *.example.org
    ControlMaster auto
    ControlPath ~/.ssh/cm-%C
    ControlPersist 10m
```

Later sessions to the same host share one connection. `ssh -O exit build` closes it.

### Keep idle sessions alive {#keepalive}
```text
Host *
    ServerAliveInterval 60
```

### Copy files {.note #copy}
```sh
scp -P 2222 report.pdf ada@build.example.org:/tmp/
scp -r site/ build:/srv/www/
```

`scp` uses the same keys and config; `-P` (capital) sets the port.
