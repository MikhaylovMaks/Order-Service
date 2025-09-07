package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"github.com/stretchr/testify/mock"
)

// BenchmarkGetOrder_NoCache — бенчмарк для проверки скорости получения заказа,
// когда кэш пустой. В этом случае запрос идёт в репозиторий (эмуляция задержки).
func BenchmarkGetOrder_NoCache(b *testing.B) {
	cache := storage.NewMemoryStorage()
	repo := new(mockRepo)

	repo.
		On("GetOrderByUID", mock.Anything, "slow").
		Return(&models.Order{OrderUID: "slow"}, nil).
		Run(func(args mock.Arguments) {
			time.Sleep(5 * time.Millisecond)
		})

	server := newTestServer(repo, cache)
	req := httptest.NewRequest(http.MethodGet, "/orders/slow", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.Router().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status %d", w.Code)
		}
	}
}

func BenchmarkGetOrder_WithCache(b *testing.B) {
	cache := storage.NewMemoryStorage()
	repo := new(mockRepo)

	repo.
		On("GetOrderByUID", mock.Anything, "fast").
		Return(&models.Order{OrderUID: "fast"}, nil)

	server := newTestServer(repo, cache)
	req := httptest.NewRequest(http.MethodGet, "/orders/fast", nil)

	// прогреваем кэш
	w := httptest.NewRecorder()
	server.Router().ServeHTTP(w, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.Router().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status %d", w.Code)
		}
	}
}
