include .env

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

build:
	docker compose build

## up: start all services and run migrations
up:
	docker compose up -d

## down: stop and remove all containers
down:
	docker compose down

## logs: tail logs for all services
logs:
	docker compose logs -f

## migrate/new name=$1: create a new migration file (e.g. make migrate/new name=add_blacklist_table)
migrate/new:
	@read -p "Enter migration name: " name; \
	docker run --user $(shell id -u):$(shell id -g) \
		-v $(shell pwd)/migrations:/migrations \
		migrate/migrate \
		create -ext sql -dir /migrations/ -seq $$name

## migrate/up: run all up migrations
migrate/up:
	docker compose run --rm -e POSTGRESQL_DSN=$(POSTGRESQL_DSN) migrate -path=/migrations/ -database=$(POSTGRESQL_DSN) up

## migrate/down: rollback the last migration
migrate/down:
	docker compose run --rm -e POSTGRESQL_DSN=$(POSTGRESQL_DSN) migrate -path=/migrations/ -database=$(POSTGRESQL_DSN) down 1

migrate/force:
	@read -p "Enter version to force: " v; \
	docker compose run --rm -e POSTGRESQL_DSN=$(POSTGRESQL_DSN) migrate -path=/migrations/ -database=$(POSTGRESQL_DSN) force $$v


## redis/cli: enter redis cli
redis/cli:
	docker exec -it redis redis-cli -a $(REDIS_AUTH)