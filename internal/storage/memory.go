package storage

import (
	"sync"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
)

type Cache interface {
	Get(orderUID string) (*models.Order, bool)
	Set(orderUID string, order *models.Order)
	Invalidate(orderUID string)
	InvalidateAll()
}

type MemoryStorage struct {
	mu     sync.RWMutex
	orders map[string]*models.Order
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		orders: make(map[string]*models.Order),
	}
}

func (s *MemoryStorage) Get(orderUID string) (*models.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[orderUID]
	return order, ok
}

func (s *MemoryStorage) Set(orderUID string, order *models.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[orderUID] = order
}

func (s *MemoryStorage) Invalidate(orderUID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.orders, orderUID)
}

func (s *MemoryStorage) InvalidateAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders = make(map[string]*models.Order)
}
