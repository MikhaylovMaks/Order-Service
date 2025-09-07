package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) SaveOrder(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockRepo) GetOrderByUID(ctx context.Context, uid string) (*models.Order, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *mockRepo) GetAllOrderUIDs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func newTestServer(repo postgres.OrderRepository, cache storage.Cache) *Server {
	logger, _ := zap.NewDevelopment()
	return NewServer(0, repo, cache, logger.Sugar())
}

func TestGetOrder_FromCache(t *testing.T) {
	cache := storage.NewMemoryStorage()
	expected := &models.Order{OrderUID: "abc"}
	cache.Set("abc", expected)

	repo := new(mockRepo)
	server := newTestServer(repo, cache)

	req := httptest.NewRequest(http.MethodGet, "/orders/abc", nil)
	w := httptest.NewRecorder()
	server.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got models.Order
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, expected.OrderUID, got.OrderUID)
}

func TestGetOrder_FromRepo(t *testing.T) {
	cache := storage.NewMemoryStorage()
	repo := new(mockRepo)

	expected := &models.Order{OrderUID: "xyz"}
	repo.On("GetOrderByUID", mock.Anything, "xyz").Return(expected, nil)

	server := newTestServer(repo, cache)

	req := httptest.NewRequest(http.MethodGet, "/orders/xyz", nil)
	w := httptest.NewRecorder()
	server.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got models.Order
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, expected.OrderUID, got.OrderUID)

	// Проверим, что заказ закэширован
	cached, ok := cache.Get("xyz")
	assert.True(t, ok)
	assert.Equal(t, expected, cached)
}

func TestGetOrder_NotFound(t *testing.T) {
	cache := storage.NewMemoryStorage()
	repo := new(mockRepo)
	repo.On("GetOrderByUID", mock.Anything, "notfound").Return(nil, postgres.ErrOrderNotFound)

	server := newTestServer(repo, cache)

	req := httptest.NewRequest(http.MethodGet, "/orders/notfound", nil)
	w := httptest.NewRecorder()
	server.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetOrder_BadRequest(t *testing.T) {
	cache := storage.NewMemoryStorage()
	repo := new(mockRepo)

	server := newTestServer(repo, cache)

	req := httptest.NewRequest(http.MethodGet, "/orders/", nil) // без order_uid
	w := httptest.NewRecorder()
	server.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code) // mux вернёт 404
}
