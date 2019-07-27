.PHONY: build
build:
	bash cross-platform-build.sh github.com/zwhitchcox/zetup

.PHONY: run
run:
	go run *.go
