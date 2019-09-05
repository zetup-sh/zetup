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