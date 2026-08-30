#!/usr/bin/env bash
# Sets up the Pi relay: Syncthing + File Browser + Telegram relaybot.
# Run as your normal user (not root) on Raspberry Pi OS Lite 64-bit.
set -euo pipefail

if [[ $EUID -eq 0 ]]; then echo "run as your normal user, not root"; exit 1; fi

read -rsp "Telegram BOT_TOKEN (empty = skip relaybot): " BOT_TOKEN </dev/tty; echo

sudo apt update
sudo apt install -y syncthing curl

mkdir -p "$HOME/PixelBackup" "$HOME/filebrowser-data"

sudo systemctl enable --now "syncthing@$USER"

command -v filebrowser >/dev/null ||
  curl -fsSL https://raw.githubusercontent.com/filebrowser/get/master/get.sh | bash

sudo tee /etc/systemd/system/filebrowser.service >/dev/null <<EOF
[Unit]
Description=File Browser
After=network.target

[Service]
ExecStart=/usr/local/bin/filebrowser -r $HOME/PixelBackup -a 0.0.0.0 -p 8080 -d $HOME/filebrowser-data/filebrowser.db
Restart=always
User=$USER

[Install]
WantedBy=multi-user.target
EOF

if [[ -n $BOT_TOKEN ]]; then
  sudo curl -fsSL -o /usr/local/bin/relaybot \
    https://github.com/4n4k1n/PixelVault/releases/download/latest/relaybot
  sudo chmod +x /usr/local/bin/relaybot

  printf 'BOT_TOKEN=%s\n' "$BOT_TOKEN" | sudo tee /etc/relaybot.env >/dev/null
  sudo chmod 600 /etc/relaybot.env

  sudo tee /etc/systemd/system/relaybot.service >/dev/null <<EOF
[Unit]
Description=Pixel Vault Telegram relay
After=network-online.target

[Service]
EnvironmentFile=/etc/relaybot.env
ExecStart=/usr/local/bin/relaybot
Restart=always
User=$USER

[Install]
WantedBy=multi-user.target
EOF
fi

sudo systemctl daemon-reload
sudo systemctl enable --now filebrowser
if [[ -n $BOT_TOKEN ]]; then sudo systemctl enable --now relaybot; fi

cat <<EOF

Done.
  File Browser : http://$(hostname -I | awk '{print $1}'):8080  (login admin/admin - change it)
  Syncthing GUI: ssh -L 8384:localhost:8384 $USER@$(hostname).local  then http://localhost:8384
  Next         : create ~/PixelBackup/<name>, share it to the Pixel (README steps 4-6)
EOF
