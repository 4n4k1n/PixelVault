# Pixel Photo Relay

Use an original Google Pixel/Pixel XL (2016) to get free, unlimited, original-quality Google Photos backup — fed automatically from a Raspberry Pi.

## Why

The original Pixel and Pixel XL are the only phones with an unlimited, no-expiration, original-quality Google Photos backup perk. This project turns a Raspberry Pi into an automated relay: drop files on the Pi, they sync to the Pixel, the Pixel backs them up to Google Photos at full quality.

## How it works

```
Your files → Raspberry Pi → Syncthing → Pixel (DCIM/<name>) → Google Photos backup
```

## Requirements

- Raspberry Pi (3B+ or better) with Raspberry Pi OS Lite (64-bit)
- Original Google Pixel or Pixel XL, per Google account you want to use
- Syncthing-Fork (F-Droid) on the Pixel
- A Google account per person/phone slot

## Setup

### 1. Flash the Pi

Raspberry Pi Imager → Raspberry Pi OS Lite (64-bit) → set hostname, enable SSH, configure Wi-Fi in advanced settings. Reserve a static IP for the Pi in your router.

### 2. Run the setup script

```
ssh pi@<hostname>.local
sudo apt full-upgrade -y
curl -fsSL https://raw.githubusercontent.com/4n4k1n/PixelVault/main/setup.sh | bash
```

It asks for a Telegram bot token (leave empty to skip the bot), then installs and enables:

- **Syncthing** — the sync engine
- **File Browser** on port 8080 — web upload UI, default login `admin`/`admin`, change it
- **relaybot** — Telegram upload bot, if you gave a token

Re-running it is safe.

### 3. Create a folder per person

Open the Syncthing GUI:

```
ssh -L 8384:localhost:8384 pi@<hostname>.local
```

Then `http://localhost:8384` → add folder `~/PixelBackup/<name>`, **Folder Type = Send & Receive**, **File Versioning = Trash Can** (clean out after 14 days) — when the phone auto-deletes a backed-up file, the delete syncs back and cleans up the Pi too.

### 4. Pair the Pixel

Install Syncthing-Fork (F-Droid), add the Pi as a device, accept the folder share, and set the local folder to:

```
/storage/emulated/0/DCIM/<name>
```

**Important:** it must be under `DCIM`, not a separate folder — Google Photos only auto-backs-up folders inside `DCIM`.

### 5. Turn on Google Photos backup

On the Pixel, per account: Google Photos → Backup → turn on. Original quality is automatic on this device (no manual quality toggle needed).

### 6. Get files onto the Pi

Three ways, pick any:

- **File Browser** — `http://<pi-ip>:8080`, drag and drop.
- **Telegram bot** — send files to your bot. Must be sent as **File** (attach → File), not as a photo, or Telegram compresses them. Bots can only download up to 20 MB. Who may upload is set in `/etc/relaybot.env` as `USERS="<telegram-id>:<folder>"`, comma-separated for more people — `setup.sh` asks for it. Get your ID from `@userinfobot`. To change it later, edit that file and `sudo systemctl restart relaybot`.
- **Syncthing on your own phone** — add the Pi as a device and share the same `PixelBackup/<name>` folder. Set the phone side to **Send Only** — otherwise the auto-deletions from step 7 wipe the files on your phone too.

### 7. Manage phone storage

Enable **Smart Storage** on the Pixel (Settings → Storage). Backed-up photos are auto-deleted after 30 days, and the deletion syncs back to the Pi.

### 8. Remote access (optional)

Tailscale puts the Pi and your phone on one private encrypted network, so File Browser, SSH and the Syncthing GUI reach the Pi from mobile data as if you were home. The photo sync itself does **not** need it.

On the Pi:

```
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up          # open the login link it prints
tailscale ip -4            # the 100.x.x.x address to use
```

Install the Tailscale app on your phone and log in with the same account. Then File Browser is `http://100.x.x.x:8080` and SSH goes to the same IP.

## Files in this repo

- `setup.sh` — installs and enables everything on the Pi (step 2)
- `main.go` — the Telegram relay bot
- `.github/workflows/release.yml` — builds the arm64 bot binary on push and republishes the `latest` release, which `setup.sh` downloads

## Notes

- Works identically for photos and videos.
- The Pi is a **relay, not a permanent archive** — Google Photos is the actual long-term storage.
- There is no API to verify backup status. Instead, Google Photos' own auto-delete confirms it: a file deleted from the phone was backed up. Files older than ~30 days on the Pi mean backup is stalled — check the phone.
