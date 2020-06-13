#!/bin/sh
echo "Copying binary to /usr/sbin..."
sudo cp thinkbacklight /usr/sbin
echo "Copying configfile to /etc/thinkbacklight..."
sudo mkdir -p /etc/thinkbacklight
sudo cp config.yaml /etc/thinkbacklight
echo "Copying service object to /etc/systemd/system..."
sudo cp thinkbacklight.service  /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable thinkbacklight.service
sudo systemctl stop thinkbacklight.service
sudo systemctl start thinkbacklight.service