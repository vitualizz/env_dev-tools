#!/usr/bin/env bash

# Install base
echo "========== Install resources for install libraries =========="
apt-get install -y curl \
                   git \
                   software-properties-common \
                   gnupg-agent \
                   ca-certificates \
                   apt-transport-https \

# Docker
echo "========== Import Dependency Docker =========="
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
apt-key fingerprint 0EBFCD88
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

# Spotify
echo "========== Import Dependency Spotify =========="
curl -sS https://download.spotify.com/debian/pubkey.gpg | sudo apt-key add - 
add-apt-repository "deb http://repository.spotify.com stable non-free"

# Update & Upgrade
echo "========== Update && Upgrade =========="
apt-get update && apt-get upgrade
