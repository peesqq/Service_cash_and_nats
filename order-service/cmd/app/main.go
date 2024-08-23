package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/b0pof/order-service/internal/cache"
	"github.com/b0pof/order-service/internal/consumer"
	"github.com/b0pof/order-service/internal/handler"
	"github.com/b0pof/order-service/pkg/postgres"
	"github.com/gorilla/mux"
)

const ServerPort = ":8000"
const CacheTTL = 3 * time.Minute
const OrderSubject = "orders"

func main() {
	ctx := context.TODO()

	// db
	dbConn, err := postgres.Connect(ctx)
	if err != nil {
		log.Fatalf("cant' connect to postgres: %s", err)
	}

	// cache
	c := cache.NewOrderCache(CacheTTL, dbConn)

	// nats
	cons, err := consumer.NewConsumer()
	if err != nil {
		log.Fatalf("cant' connect to nats: %s", err)
	}
	defer cons.Close()
	if err = cons.Subscribe(OrderSubject, c, dbConn); err != nil {
		log.Fatalf("cant' subscribe: %s", err)
	}

	// handler
	h := handler.New(dbConn, c)

	// router
	r := mux.NewRouter()
	r.HandleFunc("/api/order/{id}", h.GetOrder).Methods("GET")

	// server
	srv := http.Server{
		Addr:    ServerPort,
		Handler: r,
	}

	log.Println("Starting server on " + ServerPort)
	if err = srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
