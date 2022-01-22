# Film API

## Development

Build and run app on port :4000.

```bash
go build -o ./bin ./...
./bin/api
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