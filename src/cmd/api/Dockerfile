FROM golang:1.25-alpine AS build

WORKDIR /app

COPY . /app

RUN go build -o /app/api_service /app/src/cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/api_service /app/.env /app/.env.secret /app/

EXPOSE 8080

ENTRYPOINT ["/app/api_service"]
