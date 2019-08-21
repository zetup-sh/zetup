.PHONY: build
build:
	rm -rf ./build/*
	bash ./scripts/cross-platform-build.sh github.com/zetup-sh/zetup


.PHONY: release
release:
	python ./scripts/release.py

.PHONY: push-test
push-test: build
	scp -r ./build/* 192.168.1.68:/var/www/zetup.sh/html/test