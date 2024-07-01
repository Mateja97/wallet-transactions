package nats_server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"sync"
	"transactions/models"
	"transactions/shared"
)

type BalanceResponse struct {
	ID        string  `json:"id"`
	Balance   float64 `json:"balance"`
	RequestId string  `json:"request_id"`
}

type BalanceRequest struct {
	ID string `json:"id"`
}

func BalanceSubscribe(wg *sync.WaitGroup) error {
	log.Println("balance request subscribed")
	defer wg.Done()

	sub, err := n.natsConn.SubscribeSync("balance-request")
	if err != nil {
		log.Printf("subscribe failed: %v", err)
		return err
	}

	defer func(_sub *nats.Subscription) {
		err = _sub.Unsubscribe()
		if err != nil {
			log.Printf("sub unsubscribe has failed: %v", err)
		}
	}(sub)

	for {
		m, err := sub.NextMsg(nats.DefaultTimeout)
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			}
			log.Printf("error receiving nats message: %v", err)
			continue
		}
		var request BalanceRequest
		err = json.Unmarshal(m.Data, &request)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			continue
		}

		userLock := shared.Resources.GetUserLock(request.ID)
		userLock.Lock()
		var user models.DBUser
		n.db.Where(&models.DBUser{UserID: request.ID}).First(&user)
		userLock.Unlock()

		response, err := json.Marshal(BalanceResponse{
			ID:      request.ID,
			Balance: user.Balance,
		})
		if err != nil {
			log.Printf("Error marshalling response: %v", err)
			continue
		}
		err = n.natsConn.Publish(m.Reply, response)
		if err != nil {
			log.Printf("Error message publish: %v", err)
			continue
		}
		log.Printf("Response has been sent for user id %s:", request.ID)
	}

}
