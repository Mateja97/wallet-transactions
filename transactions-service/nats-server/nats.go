package nats_server

import (
	"github.com/jinzhu/gorm"
	"github.com/nats-io/nats.go"
	"sync"
)

type NatsServer struct {
	db       *gorm.DB
	natsConn *nats.Conn
	mu       *sync.Mutex
}

var n *NatsServer

func Init(opts ...func(*NatsServer) error) error {
	n = &NatsServer{}

	for _, o := range opts {
		err := o(n)
		if err != nil {
			return err
		}
	}

	return nil
}

func Shutdown() {
	n.natsConn.Close()
}
