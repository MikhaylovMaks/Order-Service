package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

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
	return &Server{
		port:  port,
		repo:  repo,
		cache: cache,
		log:   log}
}

func (s *Server) Router() http.Handler {
	r := mux.NewRouter()

	// middlewares
	r.Use(withRequestID)
	r.Use(s.withRecovery)
	r.Use(s.withLogging)
	r.Use(withTimeout(15 * time.Second))

	// health
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	// API
	r.HandleFunc("/orders/{order_uid}", s.GetOrder).Methods(http.MethodGet)

	// Static files
	webDir := filepath.Clean("./web")
	fs := http.FileServer(http.Dir(webDir))
	r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

	return r
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.srv = &http.Server{
		Addr:    addr,
		Handler: s.Router(),
	}
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

// middlewares

type ctxKey string

const ctxKeyReqID ctxKey = "req_id"

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), ctxKeyReqID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(sw, r)
		rid, _ := r.Context().Value(ctxKeyReqID).(string)
		s.log.Infow("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.code,
			"dur_ms", time.Since(start).Milliseconds(),
			"req_id", rid,
		)
	})
}

func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				rid, _ := r.Context().Value(ctxKeyReqID).(string)
				s.log.Errorw("panic", "err", rec, "req_id", rid)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func withTimeout(d time.Duration) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, "request timeout")
	}
}
