#!/bin/bash
echo "Repos"
echo "Add spotify"
curl -sS https://download.spotify.com/debian/pubkey.gpg | sudo apt-key add - 
echo "deb http://repository.spotify.com stable non-free" | sudo tee /etc/apt/sources.list.d/spotify.list

echo "Update"
apt-get update

echo "Preload"
apt install -y preload

echo "Gnome"
apt install -y gnome-tweaks

echo "Install GIT"
apt install -y git

echo "Install curl"
apt install -y curl

echo "Install ZSH"
apt install -y zsh

echo "Install Terminator"
apt install -y terminator

echo "Install neovim"
apt install -y neovim

echo "Neovim Plugins"
curl -fLo ~/.local/share/nvim/site/autoload/plug.vim --create-dirs https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim

echo "Install Docker"
apt-get install apt-transport-https ca-certificates gnupg-agent software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
apt-key fingerprint 0EBFCD88
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu disco stable"
apt install -y docker-ce
usermod -aG docker $USER

echo "Install Docker Compose"
curl -L "https://github.com/docker/compose/releases/download/1.25.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

echo "Install Spotify"
apt install -y spotify-client
