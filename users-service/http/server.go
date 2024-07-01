package http

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/nats-io/nats.go"
	"github.com/segmentio/kafka-go"
	"net/http"
	"sync"
)

type Server struct {
	db          *gorm.DB
	kafkaWriter *kafka.Writer

	mu         sync.Mutex
	httpServer *http.Server
	natsConn   *nats.Conn
}

var srv *Server

func Init(opts ...func(*Server) error) error {
	srv = &Server{}

	for _, o := range opts {
		err := o(srv)
		if err != nil {
			return err
		}
	}

	return nil
}
func Start() error {
	return srv.httpServer.ListenAndServe()
}

func Shutdown(ctx context.Context) error {
	return srv.httpServer.Shutdown(ctx)
}
