package nats_server

import (
	"github.com/jinzhu/gorm"
	"github.com/nats-io/nats.go"
)

func DB(db *gorm.DB) func(*NatsServer) error {
	return func(*NatsServer) error {
		n.db = db
		return nil
	}
}

func NatsConn(nc *nats.Conn) func(*NatsServer) error {
	return func(*NatsServer) error {
		n.natsConn = nc
		return nil
	}
}
