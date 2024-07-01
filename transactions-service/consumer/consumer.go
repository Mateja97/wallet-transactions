package consumer

import (
	"context"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/segmentio/kafka-go"

	"log"
	"time"
	"transactions/models"
)

type Consumer struct {
	db          *gorm.DB
	kafkaReader *kafka.Reader
}

const groupID = "transaction-service"

var c *Consumer

func Init(opts ...func(*Consumer) error) error {
	c = &Consumer{}

	for _, o := range opts {
		err := o(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func Consume() error {
	for {
		var m kafka.Message
		var err error
		maxRetries := 10
		retryDelay := 5 * time.Second

		for i := 0; i < maxRetries; i++ {
			m, err = c.kafkaReader.FetchMessage(context.Background())
			if err == nil {
				break
			}
			log.Printf("failed to read message, attempt %d: %v", i+1, err)
			time.Sleep(retryDelay)
		}

		if err != nil {
			log.Printf("failed to read message after %d attempts: %v", maxRetries, err)
			return err
		}
		var kafkaUser models.KafkaUser
		err = json.Unmarshal(m.Value, &kafkaUser)
		if err != nil {
			log.Printf("failed to unmarshall message %v", err)
			return err

		}
		user := &models.DBUser{
			UserID:  kafkaUser.ID,
			Balance: 0,
		}

		err = c.db.Create(user).Error
		if err != nil {
			log.Printf("error db create: %v", err)
			continue
		}

		err = c.kafkaReader.CommitMessages(context.Background(), m)
		if err != nil {
			log.Printf("failed to commit message: %v", err)
		}

	}
}

func Close() error {
	if err := c.kafkaReader.Close(); err != nil {
		log.Printf("failed to close kafka reader: %v", err)
		return err
	}
	return nil
}
