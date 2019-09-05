#!/bin/bash

VM_PASS="$1"
chmod +x /tmp/zetup
bin_dir="/usr/local/bin"
echo "$VM_PASS" | sudo -S mkdir -p "$bin_dir"
sudo mv -f /tmp/zetup "$bin_dir/zetup"
echo "$HOME/.bashrc" >> "" # touch doesn't work for some reason
if cat ~/.bashrc | grep -qv $bin_dir; then
  echo 'export PATH="${PATH}:/usr/local/bin"' >> "$HOME/.bashrc"
fi