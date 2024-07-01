package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nats-io/nats.go"
	_ "github.com/nats-io/nats.go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"transactions/consumer"
	server "transactions/http"
	"transactions/models"
	nats_server "transactions/nats-server"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("Transactions service")

	dbHost := flag.String("db-host", "", "Database host")
	dbPort := flag.Int("db-port", 5432, "Database port")
	dbUser := flag.String("db-user", "", "Database transactions-service")
	dbName := flag.String("db-name", "", "Database name")
	dbPassword := flag.String("db-password", "", "Database password")
	httpAddress := flag.String("http-address", ":8081", "HTTP server address")
	natsHost := flag.String("nats-host", "", "nats host")

	kafkaHost := flag.String("kafka-host", "kafka:9092,kafka:9093", "Kafka host")
	sourceTopic := flag.String("source-topic", "", "destination topic")
	flag.Parse()

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", *dbHost, *dbPort, *dbUser, *dbName, *dbPassword)

	log.Println(dsn)
	var err error
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(&models.DBTransaction{}, &models.DBUser{})
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

	err = consumer.Init(
		consumer.DB(db),
		consumer.KafkaReader(*kafkaHost, *sourceTopic),
	)
	if err != nil {
		log.Fatalf("Failed init server: %v", err)
	}
	go func() {
		if err := consumer.Consume(); err != nil {
			log.Fatalf("Could not consume on %s: %v\n", *kafkaHost, err)
		}
	}()

	err = server.Init(
		server.DB(db),
		server.NatsConn(natsConn),
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

	err = nats_server.Init(
		nats_server.DB(db),
		nats_server.NatsConn(natsConn),
	)
	if err != nil {
		log.Fatalf("Failed init nats server: %v", err)
	}
	wg := &sync.WaitGroup{}

	go func(_wg *sync.WaitGroup) {
		_wg.Add(1)
		err := nats_server.BalanceSubscribe(_wg)
		if err != nil {
			{
				log.Fatalf("Balance subscribe failed: %v", err)
			}
		}

	}(wg)

	wg.Wait()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("Service is shutting down...")

	err = consumer.Close()
	if err != nil {
		log.Println("consumer has failed to close")
		return
	}

	nats_server.Shutdown()
	log.Println("Nats server stopped")
	err = server.Shutdown(ctx)
	if err != nil {
		log.Println("Server has failed to shut down")
		return
	}

	log.Println("Server is shutting down...")

}
