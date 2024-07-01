package http

import (
	"github.com/jinzhu/gorm"
	"github.com/nats-io/nats.go"
	"github.com/segmentio/kafka-go"
	"net/http"
	"strings"
)

func DB(db *gorm.DB) func(*Server) error {
	return func(*Server) error {
		srv.db = db
		return nil
	}
}
func NatsConn(nc *nats.Conn) func(*Server) error {
	return func(*Server) error {
		srv.natsConn = nc
		return nil
	}
}
func KafkaWriter(kafkaHost string, destinationTopic string) func(*Server) error {
	return func(*Server) error {
		brokerList := strings.Split(kafkaHost, ",")
		srv.kafkaWriter = &kafka.Writer{
			Addr:         kafka.TCP(brokerList...),
			Topic:        destinationTopic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
		}

		return nil
	}
}
func Handler(address string) func(*Server) error {
	return func(*Server) error {
		mux := http.NewServeMux()
		mux.HandleFunc("/createUser", srv.createUser)
		mux.HandleFunc("/getBalance", srv.getBalance)

		srv.httpServer = &http.Server{
			Addr:    address,
			Handler: mux,
		}
		return nil
	}
}
