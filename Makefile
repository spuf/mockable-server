e2e:
	docker-compose -f docker-compose.test.yml build
	docker-compose -f docker-compose.test.yml run --rm sut

.PHONY: e2e
