package http

import (
	"github.com/jinzhu/gorm"
	"github.com/nats-io/nats.go"
	"net/http"
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
func Handler(address string) func(*Server) error {
	return func(*Server) error {
		mux := http.NewServeMux()
		mux.HandleFunc("/addMoney", srv.addMoney)
		mux.HandleFunc("/transferMoney", srv.transferMoney)

		srv.httpServer = &http.Server{
			Addr:    address,
			Handler: mux,
		}
		return nil
	}
}
