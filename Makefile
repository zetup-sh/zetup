.PHONY: build
build:
	rm -rf ./build/*
	bash ./scripts/cross-platform-build.sh github.com/zetup-sh/zetup

.PHONY: run
run:
	go run *.go

.PHONY: release
release:
	python ./scripts/release.py