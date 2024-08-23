package consumer

import (
	"encoding/json"
	"log"

	"github.com/b0pof/order-service/internal/cache"
	"github.com/b0pof/order-service/internal/models"
	"github.com/jmoiron/sqlx"
	stan "github.com/nats-io/stan.go"
)

const (
	clusterID = "test-cluster"
	clientID  = "test-client"
	natsURL   = "nats://localhost:4222"
)

type Consumer struct {
	conn stan.Conn
}

func NewConsumer() (*Consumer, error) {
	conn, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		return nil, err
	}
	return &Consumer{
		conn: conn,
	}, nil
}

func (c *Consumer) Send(subject string, msg []byte) error {
	if err := c.conn.Publish(subject, msg); err != nil {
		return err
	}
	return nil
}

func (c *Consumer) Subscribe(topic string, cache *cache.OrderCache, db *sqlx.DB) error {
	_, err := c.conn.Subscribe(topic, func(m *stan.Msg) {

		var order models.Order
		if err := json.Unmarshal(m.Data, &order); err != nil {
			log.Printf("Error unmarshalling order: %v", err)
		}

		// db
		q := `INSERT INTO "order" (user_id, title) VALUES ($1, $2) RETURNING id;`
		var orderID int
		err := db.QueryRow(q, order.UserID, order.Title).Scan(&orderID)
		if err != nil {
			log.Printf("Error inserting order: %v", err)
		}
		log.Println("Last insert id:", orderID)

		// cache
		cache.Set(orderID, order)

	})
	return err
}

func (c *Consumer) Close() {
	_ = c.conn.Close()
}
