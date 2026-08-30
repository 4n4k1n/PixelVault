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

Raspberry Pi Imager → Raspberry Pi OS Lite (64-bit) → set hostname, enable SSH, configure Wi-Fi in advanced settings.

### 2. First boot

```
ssh pi@<hostname>.local
sudo apt update && sudo apt full-upgrade -y
```

Reserve a static IP for the Pi in your router.

### 3. Install Syncthing

```
sudo apt install syncthing -y
sudo systemctl enable --now syncthing@<your-username>
```

Access the GUI via SSH tunnel:

```
ssh -L 8384:localhost:8384 pi@<hostname>.local
```

Then open `http://localhost:8384`.

### 4. Create a folder per person

e.g. `~/PixelBackup/<name>`. Set **Folder Type = Send & Receive** on the Pi side and turn on **File Versioning = Trash Can** (clean out after 14 days) — when the phone auto-deletes a backed-up file, the delete syncs back and cleans up the Pi too.

### 5. Pair the Pixel

Install Syncthing-Fork (F-Droid), add the Pi as a device, accept the folder share, and set the local folder to:

```
/storage/emulated/0/DCIM/<name>
```

**Important:** it must be under `DCIM`, not a separate folder — Google Photos only auto-backs-up folders inside `DCIM`.

### 6. Turn on Google Photos backup

On the Pixel, per account: Google Photos → Backup → turn on. Original quality is automatic on this device (no manual quality toggle needed).

### 7. Get files onto the Pi

Simplest: [File Browser](https://github.com/filebrowser/filebrowser), a self-hosted web upload UI:

```
curl -fsSL https://raw.githubusercontent.com/filebrowser/get/master/get.sh | bash
filebrowser -r /home/<user>/PixelBackup -a 0.0.0.0 -p 8080 -d /home/<user>/filebrowser-data/filebrowser.db
```

Run as a systemd service so it survives reboots (see `filebrowser.service` example below).

### 8. Manage phone storage

Enable **Smart Storage** on the Pixel (Settings → Storage). Backed-up photos are auto-deleted after 30 days, and the deletion syncs back to the Pi.

## Files in this repo

- `filebrowser.service` — systemd unit to keep File Browser running:

```ini
[Unit]
Description=File Browser
After=network.target
[Service]
ExecStart=/usr/local/bin/filebrowser -r /home/<user>/PixelBackup -a 0.0.0.0 -p 8080 -d /home/<user>/filebrowser-data/filebrowser.db
Restart=always
User=<user>
[Install]
WantedBy=multi-user.target
```

Enable it with:

```
sudo systemctl daemon-reload
sudo systemctl enable --now filebrowser
```

## Notes

- Works identically for photos and videos.
- The Pi is a **relay, not a permanent archive** — Google Photos is the actual long-term storage.
- There is no API to verify backup status. Instead, Google Photos' own auto-delete confirms it: a file deleted from the phone was backed up. Files older than ~30 days on the Pi mean backup is stalled — check the phone.
