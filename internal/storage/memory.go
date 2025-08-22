package storage

import (
	"sync"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
)

type Storage interface {
	Get(id string) (models.Order, bool)
	Save(order models.Order)
}

type MemoryStorage struct {
	mu     sync.RWMutex
	orders map[string]models.Order
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		orders: make(map[string]models.Order),
	}
}

func (s *MemoryStorage) Get(id string) (models.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[id]
	return order, ok
}

func (s *MemoryStorage) Save(order models.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.OrderUID] = order
}
