all: build run migrate

.env.secret:
	@if [ "$(API_KEY)" = "" ]; then \
		echo "make .env.secret: API_KEY is not set"; \
		exit 1; \
	fi
	echo "EXCHANGE_RATES_IO_API_KEY=$(API_KEY)" > .env.secret

build: build-api build-worker

build-api:
	docker build --file src/cmd/api/Dockerfile -t exchange-rates-service-api:latest .

build-worker:
	docker build --file src/cmd/worker/Dockerfile -t exchange-rates-service-worker:latest .

run:
	docker compose up -d

migrate:
	go run src/cmd/migrate/main.go

test:
	go test -v ./...

format:
	gofmt -s -w .
	goimports -w .
	go mod tidy
	swag fmt -g src/cmd/api/main.go

swagger:
	swag init -g src/cmd/api/main.go -o src/docs/

stop:
	docker compose down

clean:
	rm .env.secret
	docker rmi exchange-rates-service-api:latest
	docker rmi exchange-rates-service-worker:latest