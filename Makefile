include .envrc

current_time = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -w -X main.buildTime=${current_time} -X main.version=${git_description}'
production_host_ip = '90.132.252.1'

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


# ==================================================================================== #
# BUILD
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@echo 'Starting app on http://localhost:${PORT}'
	go run ./cmd/api -port=${PORT} -db-dsn=${FILMAPI_DB_DSN}

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

# ==================================================================================== #
# TOOLS
# ==================================================================================== #

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	golangci-lint run
	@echo 'Running tests...'
	go test -race -vet=off ./...

# ==================================================================================== #
# DATABASE
# ==================================================================================== #

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path=./migrations -database=${FILMAPI_DB_DSN} up

## db/migrations/down: revert all database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running down migrations...'
	migrate -path=./migrations -database=${FILMAPI_DB_DSN} down

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh filmapi@${production_host_ip}

## production/deploy/api: deploy the api to production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P ./bin/linux_amd64/api filmapi@${production_host_ip}:~
	rsync -rP --delete ./migrations filmapi@${production_host_ip}:~
	rsync -P ./remote/production/api.service filmapi@${production_host_ip}:~
	rsync -P ./remote/production/Caddyfile filmapi@${production_host_ip}:~
	ssh -t filmapi@${production_host_ip} '\
		migrate -path ~/migrations -database $$FILMAPI_DB_DSN up \
		&& sudo mv ~/api.service /etc/systemd/system/ \
		&& sudo systemctl enable api \
		&& sudo systemctl restart api \
		&& sudo mv ~/Caddyfile /etc/caddy/ \
		&& sudo systemctl reload caddy \
	'
