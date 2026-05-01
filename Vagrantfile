# -*- mode: ruby -*-
# vi: set ft=ruby :

# =============================================================================
# Vitualizz DevStack — Vagrantfile
# =============================================================================
# Usage:
#   vagrant up ubuntu     # Boot Ubuntu VM
#   vagrant up arch       # Boot Arch Linux VM
#   vagrant ssh ubuntu    # SSH into Ubuntu VM
#   vagrant ssh arch      # SSH into Arch Linux VM
#   vagrant destroy -f    # Destroy both VMs
#
# Requirements:
#   • VirtualBox or libvirt (Vagrant provider)
#   • Vagrant 2.3+
# =============================================================================

Vagrant.configure("2") do |config|
  # =====================================================================
  # Ubuntu (Jammy 22.04 LTS)
  # =====================================================================
  config.vm.define "ubuntu" do |u|
    u.vm.box = "ubuntu/jammy64"
    u.vm.hostname = "devstack-ubuntu"

    # 2 vCPUs, 2GB RAM — enough for all tools
    u.vm.provider "virtualbox" do |vbox|
      vbox.name = "vitualizz-devstack-ubuntu"
      vbox.cpus = 2
      vbox.memory = 2048
      vbox.default_nic_type = "virtio"
    end

    # Synced folder: project root available in VM
    u.vm.synced_folder ".", "/vagrant", disabled: false
  end

  # =====================================================================
  # Arch Linux
  # =====================================================================
  config.vm.define "arch" do |a|
    a.vm.box = "archlinux/archlinux"
    a.vm.hostname = "devstack-arch"

    a.vm.provider "virtualbox" do |vbox|
      vbox.name = "vitualizz-devstack-arch"
      vbox.cpus = 2
      vbox.memory = 2048
      vbox.default_nic_type = "virtio"
    end

    a.vm.synced_folder ".", "/vagrant", disabled: false
  end

  # =====================================================================
  # Common: Disable default /vagrant sync for Arch (pacman issues)
  # =====================================================================
  # Arch boxes sometimes have VirtualBox Guest Additions issues.
  # If sync fails, tools are still available via SSH from /vagrant.
  config.vm.synced_folder ".", "/vagrant",
    type: "rsync",
    rsync__exclude: [".git/"]
end
