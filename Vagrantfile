# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "vivid64"
  config.vm.box_url = "https://cloud-images.ubuntu.com/vagrant/vivid/current/vivid-server-cloudimg-amd64-vagrant-disk1.box"

  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.vm.synced_folder ".", "/procfs"

  config.vm.provision :shell, :inline => <<EOF
set -e

sudo apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y curl

curl -sL https://storage.googleapis.com/golang/go1.5.linux-amd64.tar.gz | tar -C /usr/local/ -zxf -
echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
source /etc/profile
EOF

end
