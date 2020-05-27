#!/usr/bin/env bash

echo "Install NeoVim"
apt install -y neovim

echo "Install Plugins for NeoVim"
curl -fLo ~/.local/share/nvim/site/autoload/plug.vim --create-dirs https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim
cp -r ./config ~/.config/nvim
