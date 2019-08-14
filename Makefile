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

.PHONY: publish-site
publish-site:
	yarn --cwd ./site build
	scp -r ./site/build/* 192.168.1.68:/var/www/zetup.sh/html