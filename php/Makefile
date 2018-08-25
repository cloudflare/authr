# dependencies
# - composer

setup: vendor/.ok

vendor/.ok:
	composer install --working-dir=../
	touch $@

clean:
	rm -rf vendor

test:
	./vendor/bin/phpunit

.PHONY: setup clean test
