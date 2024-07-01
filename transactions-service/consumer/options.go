package consumer

import (
	"github.com/jinzhu/gorm"
	"github.com/segmentio/kafka-go"
	"strings"
	"time"
)

func DB(db *gorm.DB) func(consumer *Consumer) error {
	return func(consumer *Consumer) error {
		c.db = db
		return nil
	}
}

func KafkaReader(kafkaHost string, sourceTopic string) func(consumer *Consumer) error {
	return func(consumer *Consumer) error {
		brokerList := strings.Split(kafkaHost, ",")
		c.kafkaReader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokerList,
			Topic:          sourceTopic,
			GroupID:        groupID,
			CommitInterval: time.Second,
		})

		return nil
	}
}
