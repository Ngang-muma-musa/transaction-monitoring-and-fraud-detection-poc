include .env

MIGRATE_CMD=docker compose run --rm migrate

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## up: start all services and run migrations
up:
	docker compose up -d
	@echo "Services started. Checking migrations..."

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
	$(MIGRATE_CMD) up

## migrate/down: rollback the last migration
migrate/down:
	$(MIGRATE_CMD) down 1

migrate/force:
	@read -p "Enter version to force: " v; \
	$(MIGRATE_CMD) force $$v


## redis/cli: enter redis cli
redis/cli:
	docker exec -it redis redis-cli -a $(REDIS_AUTH)