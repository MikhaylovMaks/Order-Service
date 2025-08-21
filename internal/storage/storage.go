package storage

import (
	"github.com/MikhaylovMaks/wb_techl0/internal/models"
)

type Storage interface {
	Get(id string) (models.Order, bool)
	Save(order models.Order)
}

type MemoryStorage struct {
	data map[string]models.Order
}

func (m *MemoryStorage) Get(id string) (models.Order, bool) {
	order, ok := m.data[id]
	return order, ok
}

func (m *MemoryStorage) Save(order models.Order) {
	m.data[order.OrderUID] = order
}
