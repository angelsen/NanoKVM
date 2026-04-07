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

Planned changes (not yet implemented):

- `server/service/hid/touch.go` — NEW. Multi-touch HID handler (Linux multi-touch protocol B). Models on existing `mouse.go` and `paste.go`.
- `server/service/hid/hid.go` — modified. Add `Hidg3` field + Open/Close.
- `server/router/hid.go` — modified. Register `POST /api/hid/touch_tap` route.
- `kvmapp/system/init.d/S03usbhid` — modified. Add `hid.GS3` block (multi-touch protocol B descriptor) gated on `/boot/usb.touch` flag file.
