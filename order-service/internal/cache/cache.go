package cache

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/b0pof/order-service/internal/models"
	"github.com/jmoiron/sqlx"
)

var ErrNoOrderInCache = errors.New("no order in cache")
var ErrQueryDB = errors.New("cannot query db to heat the cache")

const (
	interval = 5 * time.Second
	heatTLL  = 5 * time.Minute // TODO
)

type OrderCacheValue struct {
	order    models.Order
	cachedAt time.Time
}

type OrderCache struct {
	ttl     time.Duration
	mu      sync.RWMutex
	storage map[int]OrderCacheValue
}

func NewOrderCache(ttl time.Duration, source *sqlx.DB) *OrderCache {
	cache := &OrderCache{
		ttl:     ttl,
		storage: make(map[int]OrderCacheValue),
	}
	if err := cache.Heat(source); err != nil {
		if errors.Is(err, ErrQueryDB) {
			panic(err)
		}
	}
	go cache.Clear()
	return cache
}

func (oc *OrderCache) Heat(db *sqlx.DB) error {
	q := fmt.Sprintf("SELECT id, user_id, title, created_at FROM \"order\" WHERE created_at >= NOW() - INTERVAL '%d seconds'", int(heatTLL.Seconds()))
	rows, err := db.Query(q)
	if err != nil {
		log.Println(err)
		return ErrQueryDB
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		if err = rows.Scan(&order.ID, &order.UserID, &order.Title, &order.CreatedAt); err != nil {
			return err
		}
		oc.Set(order.ID, order)
	}
	return nil
}

func (oc *OrderCache) Clear() {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		oc.mu.Lock()
		for k, v := range oc.storage {
			if time.Now().After(v.cachedAt.Add(oc.ttl)) {
				delete(oc.storage, k)
			}
		}
		oc.mu.Unlock()
	}
}

func (oc *OrderCache) Set(id int, order models.Order) {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	oc.storage[id] = OrderCacheValue{
		order:    order,
		cachedAt: time.Now(),
	}
}

func (oc *OrderCache) Get(id int) (models.Order, error) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	value, ok := oc.storage[id]
	if !ok {
		return models.Order{}, ErrNoOrderInCache
	}
	return value.order, nil
}
