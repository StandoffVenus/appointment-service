DB_NAME=db.sqlite3
JSON_SEED_FILE=appointments.json
PORT=8080

run: seed
	go run cmd/server/main.go \
		-file "${DB_NAME}" \
		-port ${PORT}

seed:
	go run cmd/seed/main.go \
		-db "${DB_NAME}" \
		-json "${JSON_SEED_FILE}"

install:
	go build -o bin/server cmd/server/main.go
	
dump-db:
	@echo -e ".headers on\n.mode column\nSELECT * FROM appointments" | sqlite3 "${DB_NAME}"
