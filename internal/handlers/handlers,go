package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/MikhaylovMaks/wb_techl0/internal/models"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
)

type Handler struct {
	store storage.Storage
}

func NewHandler(store storage.Storage) *Handler {
	return &Handler{store: store}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

// GET /order/{id}
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/order/")

	order, ok := h.store.Get(id)
	if !ok {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// POST /order
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var order models.Order
	if err := json.Unmarshal(body, &order); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	h.store.Save(order)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("order saved"))
}
