.PHONY: build
build:
	bash cross-platform-build.sh github.com/zwhitchcox/zetup

.PHONY: run
run:
	go run *.go


.PHONY: publish-site
publish-site:
	yarn --cwd ./site build
	scp -r ./site/build/* 192.168.1.68:/var/www/zetup.sh/html