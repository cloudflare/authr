RELEASE_TYPE ?= patch

clean:
	make --directory ./js clean
	make --directory ./php clean

setup:
	make --directory ./js setup
	make --directory ./php setup
	# make --directory ./go setup (not implemented yet)

test:
	make --directory ./js test
	make --directory ./php test
	# make --directory ./go test (not implemented yet)

release:
	./contrib/semver bump $(RELEASE_TYPE) `cat VERSION` > VERSION
	make --directory ./js release
	git add VERSION js/package.json
	git commit -m "Release v`cat VERSION`"
	git tag `cat VERSION`
	git push --tags
	git push --all

.PHONY: setup clean test release
