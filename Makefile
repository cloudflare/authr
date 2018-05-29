RELEASE_TYPE ?= patch

GO_BIN := $(shell which go)

clean:
	rm -fv *.pprof *.out
	rm -rf ./vendor
	make --directory ./js clean
	make --directory ./php clean

setup:
	make --directory ./js setup
	make --directory ./php setup
	# make --directory ./go setup (not implemented yet)

test:
	$(GO_BIN) test -race .
	make --directory ./js test
	make --directory ./php test

release:
	./contrib/semver bump $(RELEASE_TYPE) `cat VERSION` > VERSION
	make --directory ./js release
	git add VERSION js/package.json
	git commit -m "Release v`cat VERSION`"
	git tag `cat VERSION`
	git push --tags
	git push --all

.PHONY: setup clean test release
