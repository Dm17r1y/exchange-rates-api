FROM golang:1.25-alpine AS build

WORKDIR /app

COPY . /app

RUN go build  -o /app/worker_service /app/src/cmd/worker/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/worker_service /app/.env /app/.env.secret /app/

ENTRYPOINT ["/app/worker_service"]
