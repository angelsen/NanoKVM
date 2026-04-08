# Upstream tracking

This is a fork of [`sipeed/NanoKVM`](https://github.com/sipeed/NanoKVM), the open-source firmware for the Sipeed NanoKVM IP-KVM hardware (RISC-V SG2002).

## Provenance

| Field | Value |
|---|---|
| Upstream | https://github.com/sipeed/NanoKVM |
| Fork | https://github.com/angelsen/NanoKVM |
| Forked at commit | `fda3c6b1a863fc33d555176e3fb0dc1f164b0708` |
| Forked on | 2026-04-07 |
| Upstream license | GPL-3.0 |
| Upstream default branch | `main` |

## Remotes

```
origin    git@github.com:angelsen/NanoKVM.git    (our fork)
upstream  https://github.com/sipeed/NanoKVM.git  (sipeed)
```

## How to update from upstream

```bash
git fetch upstream
git rebase upstream/main
git push origin main --force-with-lease
```

## Where our changes live

We follow minimal-delta discipline — prefer adding new files over modifying existing ones, cluster additions in well-named locations, no drive-by refactors. Each modification we make to upstream files is merge-friction cost.

### Multi-touch HID

- `server/service/hid/touch.go` — NEW. `TouchTap` handler (Linux multi-touch protocol B). Mirrors `paste.go`.
- `server/service/hid/hid.go` — modified. Adds `g3` file handle, `touchMutex`, `WriteHid3`, `HasTouchDevice()`. Missing `/dev/hidg3` is logged at Debug, not Error — the device is optional.
- `server/router/hid.go` — modified. Registers `POST /api/hid/touch_tap`.
- `kvmapp/system/init.d/S03usbhid` — modified. Adds the `hid.GS3` configfs block, gated on `/boot/usb.touch`. `subclass=0/protocol=0` to bind `hid-multitouch.c` rather than `hid-input.c` boot-class HID.

### Build system

- `Makefile` — modified. `DOCKER_CMD ?= docker` for podman/nerdctl substitution. `--userns=keep-id` when `DOCKER_CMD=podman` (rootless podman would otherwise bind-mount the source as a non-host UID and break `go mod tidy`). `CGO_LDFLAGS=-Wl,-rpath,$ORIGIN/dl_lib` so `libkvm.so` resolves at runtime — without this the binary has no rpath at all.
- `docker/Dockerfile` — modified. Drops the `sdk` stage (only needed for `make support`, not `make app`), trims `host_tools` to the musl toolchain only (the glibc tree is dead weight for the Go build), drops `COPY --chmod=+x` (buildah only accepts octal). Builder image: ~6-8 GB → 1.84 GB.

Four of these build fixes (`DOCKER_CMD`, `--userns`, `--chmod`, sdk-stage drop) are candidates for upstream PRs to Sipeed.
