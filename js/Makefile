V := $(shell cat ../VERSION)

prepare:
	yarn run build

setup:
	yarn

clean:
	rm -rf node_modules

test:
	yarn run test

release: prepare
	jq -M '.version = "$V"' package.json > package.json.tmp
	rm package.json
	mv package.json.tmp package.json
	npm publish

.PHONY: setup clean test prepare release
