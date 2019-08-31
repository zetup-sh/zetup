#!/bin/bash

source ./env.sh
MACHINE_NAME="Manjaro Gnome Master"

# turn off current vm and restore snapshot
# VBoxManage controlvm "${MACHINE_NAME}" poweroff
# VBoxManage snapshot "${MACHINE_NAME}" restore "SSH and Port Forwarding"
# VBoxManage startvm "${MACHINE_NAME}" --type headless

# # copy public key to vm
# cat "${HOME}\.ssh\id_rsa.pub" | ssh zwhitchcox@localhost -p 1111 "mkdir -p ~/.ssh && cat >> /home/zwhitchcox/.ssh/authorized_keys"

# build zetup and copy to vm
rm -rf ./build/*
sh ./scripts/cross-platform-build.sh linux/amd64
scp  -P 1111 -r ./build/zetup-linux-amd64  "zwhitchcox@localhost:/tmp/zetup"
ssh -t -p 1111 zwhitchcox@localhost "chmod +x /tmp/zetup && echo ${ZETUP_VM_PASS} | sudo -S mv /tmp/zetup /bin/zetup"

# run zetup on vm
ssh -t -p 1111 zwhitchcox@localhost "zetup id use gh/zwhitchcox:${GITHUB_PASS} && zetup use -b add-arch -p ssh github.com/zetup-sh/zetup-pkg"