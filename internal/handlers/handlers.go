package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/MikhaylovMaks/wb_techl0/internal/repository/postgres"
	"github.com/MikhaylovMaks/wb_techl0/internal/storage"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	port  int
	repo  postgres.OrderRepository
	cache storage.Cache
	log   *zap.SugaredLogger
	srv   *http.Server
}

func NewServer(port int, repo postgres.OrderRepository, cache storage.Cache, log *zap.SugaredLogger) *Server {
	return &Server{port: port, repo: repo, cache: cache, log: log}
}

func (s *Server) Start() error {
	r := mux.NewRouter()

	r.HandleFunc("/orders/{order_uid}", s.GetOrder).Methods(http.MethodGet)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	addr := ":" + strconv.Itoa(s.port)
	s.srv = &http.Server{Addr: addr, Handler: r}
	s.log.Infow("HTTP server listening", "addr", addr)
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Errorw("http server error", "err", err)
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	if err := s.srv.Shutdown(ctx); err != nil {
		s.log.Errorw("http server shutdown error", "err", err)
		return err
	}
	s.log.Info("HTTP server gracefully stopped")
	return nil
}

func (s *Server) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	orderUID := vars["order_uid"]
	if orderUID == "" {
		s.log.Warn("missing order_uid in path")
		http.Error(w, "missing order_uid", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if order, ok := s.cache.Get(orderUID); ok && order != nil {
		s.log.Infow("order fetched from cache", "order_uid", orderUID)
		if err := json.NewEncoder(w).Encode(order); err != nil {
			s.log.Errorw("failed to encode order (cache)", "order_uid", orderUID, "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	order, err := s.repo.GetOrderByUID(ctx, orderUID)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderNotFound) {
			s.log.Infow("order not found in db", "order_uid", orderUID)
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		s.log.Errorw("failed to fetch order from db", "order_uid", orderUID, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	s.cache.Set(orderUID, order)
	s.log.Infow("order cached", "order_uid", orderUID)

	if err := json.NewEncoder(w).Encode(order); err != nil {
		s.log.Errorw("failed to encode order (db)", "order_uid", orderUID, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
