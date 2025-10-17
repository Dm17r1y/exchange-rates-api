all: .env.secret build run migrate

format:
	gofmt -s -w .

build-api:
	docker build --file src/cmd/api/Dockerfile -t exchange-rates-service-api:latest .

build-worker:
	docker build --file src/cmd/worker/Dockerfile -t exchange-rates-service-worker:latest .

build: build-api build-worker

run:
	docker compose up -d

migrate:
	go run src/cmd/migrate/main.go

.env.secret:
	@if [ "$(API_KEY)" = "" ]; then \
		echo "API_KEY is not set"; \
		exit 1; \
	fi
	echo "EXCHANGE_RATES_IO_API_KEY=$(API_KEY)" > .env.secret

stop:
	docker compose down

clean:
	rm .env.secret