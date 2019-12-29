#!/bin/bash
echo "Update"
apt-get update

echo "Preload"
sudo apt-get install preload

echo "Gnome"
apt install gnome-tweaks

echo "Install GIT"
apt install git

echo "Install curl"
apt install curl

echo "Install ZSH"
apt install zsh

echo "Install Terminator"
apt install terminator

echo "Install neovim"
apt install neovim

echo "Neovim Plugins"
curl -fLo ~/.local/share/nvim/site/autoload/plug.vim --create-dirs https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim

echo "Install Docker"
apt-get install apt-transport-https ca-certificates gnupg-agent software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
apt-key fingerprint 0EBFCD88
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu disco stable"
apt install docker-ce
usermod -aG docker $USER

echo "Install Docker Compose"
sudo curl -L "https://github.com/docker/compose/releases/download/1.25.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

