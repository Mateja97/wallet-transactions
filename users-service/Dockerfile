FROM golang:1.22.2 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy
RUN go mod vendor

COPY . .

RUN go build -o main .

FROM golang:1.22.2

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main", "--db-host=postgres", "--db-port=5432", "--db-user=users", "--db-name=users-db", "--db-password=users", "--kafka-host=kafka:9092" ,"--nats-host=nats:6222", "--destination-topic=user-created"]
