TS_SOURCES = $(shell find src -type f -name \*.ts)

clean:
	rm -rf build
	rm -rf node_modules

node_modules/.ok: package.json package-lock.json
	npm i
	touch $@

build/.ok: node_modules/.ok $(TS_SOURCES)
	rm -rf build
	node_modules/.bin/tsc -d
	touch $@

lint: node_modules/.ok
	node_modules/.bin/tsc --noEmit

test: node_modules/.ok
	npm run test

build: build/.ok

setup: build/.ok

.PHONY: clean build lint test setup
