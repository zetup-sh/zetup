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