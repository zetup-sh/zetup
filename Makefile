.PHONY: build
build:
	bash cross-platform-build.sh github.com/zwhitchcox/zetup

.PHONY: run
run:
	go run *.go


.PHONY: publish-site
publish-site:
	yarn --cwd ./site build
	yarn --cwd ./site pub