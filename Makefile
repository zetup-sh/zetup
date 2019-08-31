.PHONY: build
build:
	rm -rf ./build/*
	bash ./scripts/cross-platform-build.sh

.PHONY: build-linux
build-linux:
	rm -rf ./build/*
	bash ./scripts/cross-platform-build.sh linux/amd64


.PHONY: release
release:
	python ./scripts/release.py

.PHONY: push-test
publish-test: build
	scp -r ./build/* 192.168.1.68:/home/zane/dev/test-files

.PHONY: push-test-linux
publish-test-linux: build-linux
	scp -r ./build/*linux* 192.168.1.68:/home/zane/dev/test-files

.PHONY: publish-test-linux-vm
publish-test-linux-vm: build-linux
	scp -i ~/.ssh/id_rsa -P 1111 -r ./build/zetup-linux-amd64  "zwhitchcox@localhost:/tmp/zetup"
	ssh -t -p 1111 zwhitchcox@localhost "chmod +x /tmp/zetup && echo ${ZETUP_VM_PASS} | sudo -S mv /tmp/zetup /bin/zetup"

.PHONY: copy-public-key-vm
copy-public-key-vm:
	cat "${HOME}\.ssh\id_rsa.pub" | ssh zwhitchcox@localhost -p 1111 "mkdir -p ~/.ssh && cat >> /home/zwhitchcox/.ssh/authorized_keys"

.PHONY: restore-snapshot
restore-snapshot:
	VBoxManage controlvm "Manjaro Gnome Master" poweroff
	VBoxManage snapshot "Manjaro Gnome Master" restore "SSH and Port Forwarding"
	VBoxManage startvm "Manjaro Gnome Master" --type headless

.PHONY: run-zetup-init-vm
run-zetup-init-vm:
	ssh -t -p 1111 zwhitchcox@localhost "zetup id use gh/zwhitchcox:${GITHUB_PASS} && zetup use -b add-arch -p ssh github.com/zetup-sh/zetup-pkg"