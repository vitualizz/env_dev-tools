#!/usr/bin/env bash

echo "========== Install Terminator =========="
mkdir ~/.config/terminator
apt install -y terminator
cp ./config ~/.config/terminator
