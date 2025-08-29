package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
)

type Server struct {
	port  int
	repo  postgres.OrderRepository
	cache storage.Cache
}

func NewServer(port int, repo postgres.OrderRepository, cache storage.Cache) *Server {
	return &Server{port: port, repo: repo, cache: cache}
}

func (s *Server) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/orders/{id}", s.getOrder).Methods("GET")

	addr := ":" + strconv.Itoa(s.port)
	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("http server error: %v", err)
	}
}

func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// check cache
	if order, ok := s.cache.Get(id); ok {
		json.NewEncoder(w).Encode(order)
		return
	}

	// check DB
	order, err := s.repo.GetOrderByUID(r.Context(), id)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	s.cache.Set(id, order)
	json.NewEncoder(w).Encode(order)
}
