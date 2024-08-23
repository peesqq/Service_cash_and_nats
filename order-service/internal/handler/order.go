package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/b0pof/order-service/internal/models"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	"github.com/b0pof/order-service/internal/cache"
)

type Handler struct {
	db    *sqlx.DB
	cache *cache.OrderCache
}

func New(db *sqlx.DB, cache *cache.OrderCache) *Handler {
	return &Handler{
		db:    db,
		cache: cache,
	}
}

// localhost:8000/api/order/{id}

var ErrNoOrderInDB = errors.New("no order in DB")

func getOrderById(db *sqlx.DB, id int) (models.Order, error) {
	query := `SELECT id, user_id, title, created_at FROM "order" WHERE id = $1`
	row := db.QueryRow(query, id)

	var order models.Order
	err := row.Scan(&order.ID, &order.UserID, &order.Title, &order.CreatedAt)
	if err != nil {
		return models.Order{}, ErrNoOrderInDB
	}
	return order, nil
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var order models.Order
	order, err = h.cache.Get(orderID)
	if err != nil {
		if errors.Is(err, cache.ErrNoOrderInCache) {
			log.Println("No data in cache")

			order, err = getOrderById(h.db, orderID)
			if err != nil {
				log.Println("No data in db")

				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("order not found"))
				return
			}
			log.Println("Data found in db")
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	bytesOrder, _ := json.Marshal(order)

	_, _ = w.Write(bytesOrder)
}
