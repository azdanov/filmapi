migrate-up:
	migrate -path=./migrations -database='postgres://filmapi:secret@localhost/filmapi?sslmode=disable' up
migrate-down:
	migrate -path=./migrations -database='postgres://filmapi:secret@localhost/filmapi?sslmode=disable' down
