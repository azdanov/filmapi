# Film API

## Development

Run `docker compose up -d` to start PostgreSQL.

Build and run app on port :4000.

```bash
go build -o=./bin/api ./cmd/api
./bin/api
```

or

```bash
make run/api
```

Install [Air](https://github.com/cosmtrek/air) for easy development and live-reloading.

```bash
air init

# adjust .air.toml config, might need to update:
# bin = "./tmp/api"
# cmd = "go build -o ./tmp/ ./..."

air
```

### Migrations

To run migrations install golang-migrate tool.

- https://github.com/golang-migrate/migrate
- https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

Examples:

```bash
migrate create -seq -ext=.sql -dir=./migrations create_movies_table
migrate create -seq -ext=.sql -dir=./migrations add_movies_check_constraints

migrate -path=./migrations -database='postgres://filmapi:secret@localhost/filmapi?sslmode=disable' up
migrate -path=./migrations -database='postgres://filmapi:secret@localhost/filmapi?sslmode=disable' down
```

or

```bash
make db/migrations/new create_movies_table
make db/migrations/up
make db/migrations/down
```

### Deployment

This will copy setup script from local machine to a virtual machine.

__Replace IP address with your VM's in commands below, Makefile and Caddyfile__

In Caddyfile instead of IP address you can use a domain name that is associated with VM's IP address. Also email will be used for TLS certificate generation.

```bash
# Copy script
rsync -rP --delete ./remote/setup root@90.132.252.1:/root

# Run script
ssh -t root@90.132.252.1 "bash /root/setup/01.sh"

# Copy executable then migrations and apply them
make production/deploy/api
```
