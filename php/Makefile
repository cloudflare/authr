setup:
	composer install --ignore-platform-reqs --no-suggest --working-dir=../

clean:
	rm -rf vendor

test:
	./vendor/bin/phpunit

.PHONY: setup clean test
