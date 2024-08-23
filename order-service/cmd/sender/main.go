package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/b0pof/order-service/internal/models"
	"github.com/nats-io/stan.go"
)

const (
	clusterID    = "test-cluster"
	clientID     = "sender-client"
	natsURL      = "nats://localhost:4222"
	orderSubject = "orders"
)

func main() {
	conn, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatal(err)
	}

	var order models.Order

	for {
		userID := rand.Intn(1024)
		order = models.Order{
			UserID: userID,
			Title:  "order" + strconv.Itoa(userID),
		}
		bytesOrder, _ := json.Marshal(order)

		if err = conn.Publish(orderSubject, bytesOrder); err != nil {
			log.Fatalf("cant' send: %s", err)
		}
		<-time.Tick(5 * time.Second)
	}
}
