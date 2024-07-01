package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	server "user/http"
	"user/models"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("User service")

	dbHost := flag.String("db-host", "", "Database host")
	dbPort := flag.Int("db-port", 5432, "Database port")
	dbUser := flag.String("db-user", "", "Database users-service")
	dbName := flag.String("db-name", "", "Database name")
	dbPassword := flag.String("db-password", "", "Database password")
	httpAddress := flag.String("http-address", ":8080", "HTTP server address")
	natsHost := flag.String("nats-host", "", "nats host")
	kafkaHost := flag.String("kafka-host", "kafka:9092,kafka:9093", "Kafka host")
	destinationTopic := flag.String("destination-topic", "", "destination topic")

	flag.Parse()

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", *dbHost, *dbPort, *dbUser, *dbName, *dbPassword)

	var err error
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db.AutoMigrate(&models.DBUser{})
	defer func(_db *gorm.DB) {
		err := _db.Close()
		if err != nil {
			_ = fmt.Errorf("db close has failed - error: %v", err)
		}
	}(db)

	natsConn, err := nats.Connect(*natsHost)
	if err != nil {
		log.Fatalf("Failed to connect to NATS server: %v", err)
	}
	defer natsConn.Close()

	err = server.Init(
		server.DB(db),
		server.NatsConn(natsConn),
		server.KafkaWriter(*kafkaHost, *destinationTopic),
		server.Handler(*httpAddress),
	)
	if err != nil {
		log.Fatalf("Failed init server: %v", err)
	}
	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not listen on %s: %v\n", *httpAddress, err)
		}
	}()

	log.Printf("Server is ready to handle requests at %s", *httpAddress)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	err = server.Shutdown(ctx)
	if err != nil {
		log.Println("Server has failed to shut down")
		return
	}

	log.Println("Server is shutting down...")

}
