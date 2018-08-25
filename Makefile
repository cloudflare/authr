RELEASE_TYPE ?= patch

GO_BIN := $(shell which go)

clean:
	rm -rf ./vendor
	make --directory ./ts clean
	make --directory ./php clean
	find . '(' -name \*.out -o -name \*.pprof ')' -exec rm -v {} +

setup:
	make --directory ./ts setup
	make --directory ./php setup

test:
	$(GO_BIN) test -race .
	make --directory ./ts test
	make --directory ./php test

release:
	./contrib/semver bump $(RELEASE_TYPE) `cat VERSION` > VERSION
	make --directory ./ts release
	git add VERSION js/package.json
	git commit -m "Release v`cat VERSION`"
	git tag `cat VERSION`
	git push --tags
	git push --all

.PHONY: setup clean test release
