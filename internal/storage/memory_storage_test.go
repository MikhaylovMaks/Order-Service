package storage

import (
	"testing"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_SetGet(t *testing.T) {
	cache := NewMemoryStorage()
	order := &models.Order{OrderUID: "123"}

	cache.Set("123", order)

	got, ok := cache.Get("123")
	assert.True(t, ok)
	assert.Equal(t, order, got)
}

func TestMemoryStorage_Invalidate(t *testing.T) {
	cache := NewMemoryStorage()
	cache.Set("123", &models.Order{OrderUID: "123"})

	cache.Invalidate("123")
	_, ok := cache.Get("123")
	assert.False(t, ok)
}

func TestMemoryStorage_InvalidateAll(t *testing.T) {
	cache := NewMemoryStorage()
	cache.Set("123", &models.Order{OrderUID: "123"})
	cache.Set("456", &models.Order{OrderUID: "456"})

	cache.InvalidateAll()
	_, ok1 := cache.Get("123")
	_, ok2 := cache.Get("456")

	assert.False(t, ok1)
	assert.False(t, ok2)
}
